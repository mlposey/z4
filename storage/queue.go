package storage

import (
	"context"
	"go.uber.org/zap"
	"time"
	"z4/telemetry"
)

type Queue struct {
	store     *SimpleStore
	feed      chan Task
	namespace string
	closed    bool
}

func New(namespace string, store *SimpleStore) *Queue {
	q := &Queue{
		store:     store,
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
		tasks, err := t.store.Get(context.Background(), TaskRange{
			Min: lastRun,
			Max: now,
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

func (t *Queue) Add(ctx context.Context, def TaskDefinition) (Task, error) {
	return t.store.Save(ctx, def)
}

func (t *Queue) Remove(ctx context.Context, task Task) error {
	return t.store.Delete(ctx, task)
}

func (t *Queue) Namespace() string {
	return t.namespace
}

func (t *Queue) Close() error {
	// TODO: Find a better way to do this.
	// This won't immediately close the feed, and it is likely the feed
	// won't close at all. This is because the goroutine blocks until
	// a consumer takes a task from the channel.
	t.closed = true
	return nil
}
