package storage

import (
	"errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"time"
)

type QueueConfig struct {
	LastRun time.Time
}

type Queue struct {
	db        *BadgerClient
	config    QueueConfig
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
	q.loadConfig()
	go q.startConfigSync()
	go q.startFeed()
	return q
}

func (t *Queue) loadConfig() {
	// TODO: Return error instead of fatal logging.
	config, err := t.db.GetConfig(t.namespace)
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			telemetry.Logger.Fatal("failed to load namespace config from database",
				zap.Error(err))
		}

		config = QueueConfig{LastRun: time.Now().Add(-time.Minute)}
		err = t.db.SaveConfig(t.namespace, config)
		if err != nil {
			telemetry.Logger.Fatal("failed to save config to database",
				zap.Error(err))
		}
	}
	t.config = config
}

func (t *Queue) startConfigSync() {
	for !t.closed {
		err := t.db.SaveConfig(t.namespace, t.config)
		if err != nil {
			telemetry.Logger.Error("failed to save config to database",
				zap.Error(err))
		}
		time.Sleep(time.Second * 5)
	}
}

func (t *Queue) startFeed() {
	delay := time.Second
	for !t.closed {
		now := time.Now()
		// TODO: Add timeout when fetching tasks from store.
		tasks, err := t.db.GetTask(TaskRange{
			Namespace: t.namespace,
			Min:       t.config.LastRun,
			Max:       now,
		})
		if len(tasks) > 0 {
			telemetry.Logger.Debug("got tasks from db", zap.Int("count", len(tasks)))
		}
		t.config.LastRun = now

		if err != nil {
			telemetry.Logger.Error("failed to fetch tasks",
				zap.Error(err))
		}

		for _, task := range tasks {
			t.feed <- task
		}

		sleep := time.Now().Sub(t.config.LastRun)
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
	return t.db.SaveTask(task)
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
