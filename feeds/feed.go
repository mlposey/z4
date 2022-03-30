package feeds

import (
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"time"
)

type Feed struct {
	tasks     *storage.TaskStore
	config    *storage.SyncedConfig
	feed      chan storage.Task
	namespace string
	closed    bool
}

func New(namespace string, db *storage.BadgerClient) *Feed {
	q := &Feed{
		tasks:     &storage.TaskStore{Client: db},
		config:    storage.NewSyncedConfig(&storage.ConfigStore{Client: db}, namespace),
		namespace: namespace,
		feed:      make(chan storage.Task),
	}
	q.config.StartSync()
	go q.startFeed()
	return q
}

func (t *Feed) startFeed() {
	delay := time.Second
	for !t.closed {
		now := time.Now()
		// TODO: Add timeout when fetching tasks from store.
		tasks, err := t.tasks.Get(storage.TaskRange{
			Namespace: t.namespace,
			Min:       t.config.C.LastRun,
			Max:       now,
		})
		if len(tasks) > 0 {
			telemetry.Logger.Debug("got tasks from DB", zap.Int("count", len(tasks)))
		}
		t.config.C.LastRun = now

		if err != nil {
			telemetry.Logger.Error("failed to fetch tasks",
				zap.Error(err))
		}

		for _, task := range tasks {
			t.feed <- task
		}

		sleep := time.Now().Sub(t.config.C.LastRun)
		if sleep < delay {
			time.Sleep(delay - sleep)
		}
	}
	close(t.feed)
}

func (t *Feed) Tasks() <-chan storage.Task {
	return t.feed
}

func (t *Feed) Add(task storage.Task) error {
	return t.tasks.Save(task)
}

func (t *Feed) Namespace() string {
	return t.namespace
}

func (t *Feed) Close() error {
	// TODO: Find a better way to do this.
	// This won't immediately close the feed, and it is likely the feed
	// won't close at all. This is because the goroutine blocks until
	// a consumer takes a storage from the channel.
	t.closed = true
	return t.config.Close()
}
