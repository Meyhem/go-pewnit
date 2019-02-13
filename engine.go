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
	WMReport = 2 << iota
)

type Engine struct {
	concurrency uint
	activeWorkers uint
}

func (ctx* Engine) worker(id uint, sink chan WorkerMessage) {
	sink <- WorkerMessage { msgType: WMExited, workerId: id }
}

func (ctx* Engine) processMessage(msg WorkerMessage) {
	switch msg.msgType {

	case WMExited:
		ctx.activeWorkers--
		logger.Info("Worker (", msg.workerId, ") exited. ", ctx.activeWorkers, " remaining.")
	case WMReport: 

	default:
		panic("Unreachable state reached. Probably a bug.")

	}
}

func (ctx* Engine) sleep() {
	time.Sleep(time.Duration(100 * time.Millisecond))
}

func NewEngine(concurrency uint) Engine {
	return Engine {
		concurrency: concurrency,
		activeWorkers: 0,
	}
}


func (ctx* Engine) Attack(u string) {

	reportChannel := make(chan WorkerMessage)
	ctx.activeWorkers = 0

	for i := uint(0); i < ctx.concurrency; i++ {
		go ctx.worker(i, reportChannel)
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
			ctx.sleep()
		}
	}

	logger.Info("Ending attack")
}