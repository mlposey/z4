package feeds

import (
	"fmt"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"github.com/segmentio/ksuid"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type TaskStream <-chan *proto.Task

// Feed provides access to a stream of tasks that are ready to be delivered.
type Feed struct {
	tasks       *storage.TaskStore
	config      *storage.SyncedConfig
	feed        chan *proto.Task
	namespace   string
	close       chan bool
	ackDeadline time.Duration
}

func New(
	namespace string,
	db *storage.BadgerClient,
	ackDeadline time.Duration,
) (*Feed, error) {
	q := &Feed{
		tasks:       storage.NewTaskStore(db),
		config:      storage.NewSyncedConfig(&storage.ConfigStore{Client: db}, namespace),
		namespace:   namespace,
		feed:        make(chan *proto.Task),
		close:       make(chan bool),
		ackDeadline: ackDeadline,
	}

	if err := q.config.StartSync(); err != nil {
		return nil, fmt.Errorf("feed creation failed due to config error: %w", err)
	}

	go q.startFeed()
	return q, nil
}

func (f *Feed) startFeed() {
	defer close(f.feed)
	telemetry.Logger.Info("feed started",
		zap.String("namespace", f.namespace))

	for {
		select {
		case <-f.close:
			return
		default:
		}

		deliveredTasks, dErr := f.getDeliveredTasks()
		undeliveredTasks, uErr := f.getUndeliveredTasks()
		if dErr != nil || uErr != nil {
			telemetry.Logger.Error("failed to fetch tasks",
				zap.Error(multierr.Combine(dErr, uErr)))
			time.Sleep(time.Second)
			continue
		}

		if len(deliveredTasks) == 0 && len(undeliveredTasks) == 0 {
			time.Sleep(time.Millisecond * 50)
			continue
		}

		if !f.handleUndeliveredTasks(undeliveredTasks) {
			return
		}
		if !f.handleDeliveredTasks(deliveredTasks) {
			return
		}
	}
}

func (f *Feed) getDeliveredTasks() ([]*proto.Task, error) {
	config := f.config.C
	lastDeliveryTime, err := ksuid.Parse(config.LastDeliveredTask)
	if err != nil {
		return nil, err
	}

	// TODO: Consider more memory efficient way to do this.
	// Maybe using iterator so we can filter in place?

	watermark := time.Now().Add(-f.ackDeadline)
	tasks, err := f.tasks.GetRange(storage.TaskRange{
		Namespace: f.namespace,
		StartID:   storage.NewTaskID(ksuid.Nil.Time()),
		EndID:     storage.NewTaskID(watermark),
	})
	if err != nil {
		return nil, err
	}

	var filtered []*proto.Task
	for _, task := range tasks {
		if f.retryTask(task, watermark, lastDeliveryTime.Time()) {
			filtered = append(filtered, task)
		}
	}
	return filtered, nil
}

func (f *Feed) retryTask(task *proto.Task, watermark, lastDelivery time.Time) bool {
	ts, err := ksuid.Parse(task.GetId())
	if err != nil {
		telemetry.Logger.Error("failed to parse task id",
			zap.Error(err))
		return false
	}

	lastRetry := task.GetLastRetry().AsTime()
	return !ts.Time().After(lastDelivery) &&
		(lastRetry.IsZero() || lastRetry.Add(f.ackDeadline).Before(watermark))
}

func (f *Feed) getUndeliveredTasks() ([]*proto.Task, error) {
	tasks, err := f.tasks.GetRange(storage.TaskRange{
		Namespace: f.namespace,
		StartID:   f.config.C.LastDeliveredTask,
		EndID:     storage.NewTaskID(time.Now()),
	})
	if err != nil {
		return nil, err
	}

	if len(tasks) > 0 {
		// Skip last delivered tasks because we already delivered it :)
		if tasks[0].GetId() == f.config.C.LastDeliveredTask {
			// TODO: Determine if we should optimize this.
			tasks = tasks[1:]
		}
	}
	return tasks, nil
}

func (f *Feed) handleDeliveredTasks(tasks []*proto.Task) bool {
	if len(tasks) == 0 {
		return true
	}

	for _, task := range tasks {
		task.LastRetry = timestamppb.New(time.Now())

		select {
		case <-f.close:
			return false

		case f.feed <- task:
		}
	}

	err := f.tasks.SaveAll(tasks)
	if err != nil {
		telemetry.Logger.Error("failed to update retry tasks",
			zap.Error(err))
		return true
	}
	return true
}

func (f *Feed) handleUndeliveredTasks(tasks []*proto.Task) bool {
	if len(tasks) == 0 {
		return true
	}

	for _, task := range tasks {
		select {
		case <-f.close:
			return false

		case f.feed <- task:
			f.config.C.LastDeliveredTask = task.GetId()
		}
	}
	return true
}

func (f *Feed) Tasks() TaskStream {
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
