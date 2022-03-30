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
	config := t.config.C

	for !t.closed {
		// TODO: Add timeout when fetching tasks from store.
		tasks, err := t.tasks.Get(storage.TaskRange{
			Namespace: t.namespace,
			StartID:   config.LastDeliveredTask,
			EndID:     storage.NewTaskID(time.Now()),
		})
		if err != nil {
			telemetry.Logger.Error("failed to fetch tasks",
				zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		if len(tasks) > 0 {
			if tasks[0].ID == config.LastDeliveredTask {
				// TODO: Determine if we should optimize this.
				tasks = tasks[1:]
			}
		}

		if len(tasks) == 0 {
			time.Sleep(time.Millisecond * 50)
			continue
		} else {
			telemetry.Logger.Debug("got tasks from DB", zap.Int("count", len(tasks)))
		}

		for _, task := range tasks {
			t.feed <- task
			t.config.C.LastDeliveredTask = task.ID
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
	// TODO: Call this method before closing app.
	// Important because failing to save the config to disk
	// could cause duplicate deliveries.

	// TODO: Find a better way to do this.
	// This won't immediately close the feed, and it is likely the feed
	// won't close at all. This is because the goroutine blocks until
	// a consumer takes a storage from the channel.
	t.closed = true
	return t.config.Close()
}
