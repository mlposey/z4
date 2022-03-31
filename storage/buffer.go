package storage

import (
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"time"
)

type taskBuffer struct {
	flushInterval time.Duration
	batchSize     int
	tasks         []Task
	idx           int
	handler       func([]Task) error
	incTasks      chan Task
	closer        chan interface{}
}

func newTaskBuffer(flushInterval time.Duration, size int, handler func([]Task) error) *taskBuffer {
	buffer := &taskBuffer{
		flushInterval: flushInterval,
		batchSize:     size,
		tasks:         make([]Task, size),
		handler:       handler,
		incTasks:      make(chan Task),
		closer:        make(chan interface{}),
	}
	go buffer.startFlushHandler()
	return buffer
}

func (tb *taskBuffer) startFlushHandler() {
	for {
		select {
		case <-tb.closer:
			var err error
			if tb.idx > 0 {
				err = tb.handler(tb.tasks[0:tb.idx])
				tb.idx = 0
			}
			tb.closer <- err
			return

		case task := <-tb.incTasks:
			tb.tasks[tb.idx] = task
			tb.idx++
			if tb.idx == len(tb.tasks) {
				err := tb.handler(tb.tasks[0:tb.idx])
				if err != nil {
					// TODO: Retry flush.
					telemetry.Logger.Error("failed to flush tasks", zap.Error(err))
				}
				tb.idx = 0
			}
		}
	}
}

func (tb *taskBuffer) Close() error {
	tb.closer <- nil
	return (<-tb.closer).(error)
}

func (tb *taskBuffer) Add(task Task) {
	// TODO: Consider performance enhancements.
	// This will block if the buffer reaches the max count.
	// We may not want to block ever; something to consider.
	tb.incTasks <- task
}
