package storage

import (
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"sync"
	"time"
)

type taskBuffer struct {
	flushInterval time.Duration
	batchSize     int
	tasks         []Task
	idx           int
	handler       func([]Task) error
	incTasks      chan Task
	closeReq      chan interface{}
	closeRes      chan error
}

func newTaskBuffer(flushInterval time.Duration, size int, handler func([]Task) error) *taskBuffer {
	buffer := &taskBuffer{
		flushInterval: flushInterval,
		batchSize:     size,
		tasks:         make([]Task, size),
		handler:       handler,
		incTasks:      make(chan Task),
		closeReq:      make(chan interface{}),
		closeRes:      make(chan error),
	}
	go buffer.startFlushHandler()
	return buffer
}

func (tb *taskBuffer) startFlushHandler() {
	var outstandingFlushes sync.WaitGroup
	for {
		select {
		case <-tb.closeReq:
			outstandingFlushes.Wait()

			var err error
			if tb.idx > 0 {
				err = tb.handler(tb.tasks[0:tb.idx])
				tb.idx = 0
			}
			tb.closeRes <- err
			return

		case task := <-tb.incTasks:
			tb.tasks[tb.idx] = task
			tb.idx++
			if tb.idx < len(tb.tasks) {
				continue
			}

			tasks := make([]Task, tb.idx)
			copy(tasks, tb.tasks[0:tb.idx])
			tb.idx = 0

			outstandingFlushes.Add(1)
			go func(t []Task) {
				defer outstandingFlushes.Done()

				err := tb.handler(t)
				if err != nil {
					// TODO: Retry flush.
					telemetry.Logger.Error("failed to flush tasks", zap.Error(err))
				}
			}(tasks)
		}
	}
}

func (tb *taskBuffer) Close() error {
	tb.closeReq <- nil
	return <-tb.closeRes
}

func (tb *taskBuffer) Add(task Task) {
	// TODO: Consider performance enhancements.
	// This will block if the buffer reaches the max count.
	// We may not want to block ever; something to consider.
	tb.incTasks <- task
}
