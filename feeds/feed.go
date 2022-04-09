package feeds

import (
	"fmt"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"time"
)

// Feed provides access to a stream of tasks that are ready to be delivered.
type Feed struct {
	tasks     *storage.TaskStore
	config    *storage.SyncedConfig
	feed      chan *proto.Task
	namespace string
	close     chan bool
}

func New(namespace string, db *storage.BadgerClient) *Feed {
	q := &Feed{
		tasks:     storage.NewTaskStore(db),
		config:    storage.NewSyncedConfig(&storage.ConfigStore{Client: db}, namespace),
		namespace: namespace,
		feed:      make(chan *proto.Task),
		close:     make(chan bool),
	}
	q.config.StartSync()
	go q.startFeed()
	return q
}

func (f *Feed) startFeed() {
	telemetry.Logger.Info("feed started",
		zap.String("namespace", f.namespace))
	config := f.config.C

LOOP:
	for {
		select {
		case <-f.close:
			break LOOP
		default:
		}

		tasks, err := f.tasks.Get(storage.TaskRange{
			Namespace: f.namespace,
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
			if tasks[0].GetId() == config.LastDeliveredTask {
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
			select {
			case f.feed <- task:
				f.config.C.LastDeliveredTask = task.GetId()

			case <-f.close:
				break LOOP
			}
		}
	}
	close(f.feed)
}

func (f *Feed) Tasks() <-chan *proto.Task {
	return f.feed
}

func (f *Feed) Namespace() string {
	return f.namespace
}

// Close stops the feed from listening to ready tasks.
//
// This method must be called once after the feed is no longer
// needed.
func (f *Feed) Close() error {
	f.close <- true
	err := f.config.Close()
	if err != nil {
		return fmt.Errorf("failed to close feed for namespace '%s': %w", f.namespace, err)
	}

	telemetry.Logger.Info("feed stopped",
		zap.String("namespace", f.namespace))
	return nil
}
