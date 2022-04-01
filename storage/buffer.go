package storage

import (
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"sync"
	"time"
)

type taskBuffer struct {
	flushInterval      time.Duration
	batchSize          int
	tasks              []*proto.Task
	idx                int
	handler            func([]*proto.Task) error
	incTasks           chan *proto.Task
	closeReq           chan interface{}
	closeRes           chan error
	outstandingFlushes sync.WaitGroup
}

func newTaskBuffer(flushInterval time.Duration, size int, handler func([]*proto.Task) error) *taskBuffer {
	buffer := &taskBuffer{
		flushInterval: flushInterval,
		batchSize:     size,
		tasks:         make([]*proto.Task, size),
		handler:       handler,
		incTasks:      make(chan *proto.Task),
		closeReq:      make(chan interface{}),
		closeRes:      make(chan error),
	}
	go buffer.startFlushHandler()
	return buffer
}

func (tb *taskBuffer) startFlushHandler() {
	interval := time.NewTicker(tb.flushInterval)
	for {
		select {
		case <-tb.closeReq:
			if tb.idx > 0 {
				tb.flush()
			}
			tb.outstandingFlushes.Wait()
			tb.closeRes <- nil
			return

		case <-interval.C:
			if tb.idx > 0 {
				tb.flush()
			}

		case task := <-tb.incTasks:
			tb.tasks[tb.idx] = task
			tb.idx++
			if tb.idx == len(tb.tasks) {
				tb.flush()
			}
			interval.Reset(tb.flushInterval)
		}
	}
}

func (tb *taskBuffer) flush() {
	tasks := make([]*proto.Task, tb.idx)
	copy(tasks, tb.tasks[0:tb.idx])
	tb.idx = 0

	tb.outstandingFlushes.Add(1)
	go func(t []*proto.Task) {
		defer tb.outstandingFlushes.Done()

		err := tb.handler(t)
		if err != nil {
			// TODO: Retry flush.
			telemetry.Logger.Error("failed to flush tasks", zap.Error(err))
		}
	}(tasks)
}

func (tb *taskBuffer) Close() error {
	tb.closeReq <- nil
	return <-tb.closeRes
}

func (tb *taskBuffer) Add(task *proto.Task) {
	// TODO: Consider performance enhancements.
	// This will block if the buffer reaches the max count.
	// We may not want to block ever; something to consider.
	tb.incTasks <- task
}
