package q

import (
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"sync"
	"time"
)

type taskWriter struct {
	tasks *storage.TaskStore
	close chan bool

	ackBuffer     []*proto.Ack
	ackMu         sync.Mutex
	flushInterval *time.Ticker
}

func NewTaskWriter(tasks *storage.TaskStore) TaskWriter {
	w := &taskWriter{
		tasks: tasks,
		close: make(chan bool),
	}
	go w.handleAckFlush()
	return w
}

func (s *taskWriter) PurgeTasks(namespace string) error {
	return s.tasks.PurgeTasks(namespace)
}

func (s *taskWriter) Push(tasks []*proto.Task, isFollower bool) error {
	return s.tasks.SaveAll(tasks, isFollower)
}

func (s *taskWriter) Acknowledge(acks []*proto.Ack) error {
	s.ackMu.Lock()
	defer s.ackMu.Unlock()

	s.ackBuffer = append(s.ackBuffer, acks...)

	if len(s.ackBuffer) > 2000 {
		s.flush()
		s.flushInterval.Reset(time.Millisecond * 500)
	}
	return nil
}

func (s *taskWriter) handleAckFlush() {
	s.flushInterval = time.NewTicker(time.Millisecond * 500)
	for {
		select {
		case <-s.close:
			return

		case <-s.flushInterval.C:
			s.ackMu.Lock()

			if len(s.ackBuffer) > 0 {
				s.flush()
			}

			s.ackMu.Unlock()
		}
	}
}

func (s *taskWriter) flush() {
	err := s.tasks.DeleteAll(s.ackBuffer)
	if err != nil {
		telemetry.Logger.Error("error acking tasks", zap.Error(err))
		return
	}
	s.ackBuffer = nil
}

func (s *taskWriter) Close() error {
	s.close <- true
	return nil
}
