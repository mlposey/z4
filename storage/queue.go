package storage

import (
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"time"
)

type Queue struct {
	db        *BadgerClient
	feed      chan Task
	namespace string
	closed    bool
}

func New(namespace string, db *BadgerClient) *Queue {
	q := &Queue{
		db:        db,
		namespace: namespace,
		feed:      make(chan Task),
	}
	go q.startFeed()
	return q
}

func (t *Queue) startFeed() {
	delay := time.Millisecond * 10
	var lastRun time.Time
	for !t.closed {
		now := time.Now()
		// TODO: Add timeout when fetching tasks from store.
		tasks, err := t.db.Get(TaskRange{
			Namespace: t.namespace,
			Min:       lastRun,
			Max:       now,
		})
		lastRun = now

		if err != nil {
			telemetry.Logger.Error("failed to fetch tasks",
				zap.Error(err))
		}

		for _, task := range tasks {
			t.feed <- task
		}

		sleep := time.Now().Sub(lastRun)
		if sleep < delay {
			time.Sleep(delay - sleep)
		}
	}
	close(t.feed)
}

func (t *Queue) Feed() <-chan Task {
	return t.feed
}

func (t *Queue) Add(task Task) error {
	return t.db.Save(t.namespace, task)
}

func (t *Queue) Namespace() string {
	return t.namespace
}

func (t *Queue) Close() error {
	// TODO: Find a better way to do this.
	// This won't immediately close the feed, and it is likely the feed
	// won't close at all. This is because the goroutine blocks until
	// a consumer takes a storage from the channel.
	t.closed = true
	return nil
}
