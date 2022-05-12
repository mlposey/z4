package q

import (
	"github.com/bluele/gcache"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"sync"
	"time"
)

type taskWriter struct {
	tasks    *storage.TaskStore
	seqCache gcache.Cache
	close    chan bool

	ackBuffer     []*proto.Ack
	ackMu         sync.Mutex
	flushInterval *time.Ticker
}

func NewTaskWriter(tasks *storage.TaskStore, namespaces *storage.NamespaceStore) TaskWriter {
	release := func(namespace string, seq *badger.Sequence) {
		telemetry.Logger.Info("closing seq")
		err := seq.Release()
		if err != nil {
			telemetry.Logger.Error("failed to close sequence",
				zap.Error(err),
				zap.String("namespace", namespace))
		}
	}

	w := &taskWriter{
		tasks: tasks,
		close: make(chan bool),
		seqCache: gcache.New(100).
			ARC().
			LoaderFunc(func(i interface{}) (interface{}, error) {
				return namespaces.Sequence(i.(string))
			}).
			EvictedFunc(func(key interface{}, value interface{}) {
				release(key.(string), value.(*badger.Sequence))
			}).
			PurgeVisitorFunc(func(key interface{}, value interface{}) {
				release(key.(string), value.(*badger.Sequence))
			}).
			Build(),
	}
	go w.handleAckFlush()
	return w
}

func (s *taskWriter) PurgeTasks(namespace string) error {
	return s.tasks.PurgeTasks(namespace)
}

func (s *taskWriter) NextIndex(namespace string) (uint64, error) {
	seq, err := s.seqCache.Get(namespace)
	if err != nil {
		return 0, err
	}
	return seq.(*badger.Sequence).Next()
}

func (s *taskWriter) Push(tasks []*proto.Task) error {
	return s.tasks.SaveAll(tasks)
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
	s.seqCache.Purge()
	return nil
}
