package sched

import (
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"io"
	"time"
)

// undeliveredTaskFetcher fetches tasks that have not been delivered to clients.
type undeliveredTaskFetcher struct {
	Tasks     *storage.TaskStore
	StartID   string
	Namespace string
}

func (utf *undeliveredTaskFetcher) Process(handle func(task *proto.Task) error) error {
	it := storage.NewTaskIterator(utf.Tasks.Client, &storage.ScheduledRange{
		Namespace: utf.Namespace,
		StartID:   utf.StartID,
		EndID:     storage.NewTaskID(time.Now()),
	})

	first, err := it.Peek()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	if first.GetId() == utf.StartID {
		// Skip last delivered tasks because we already delivered it :)
		it.Next()
	}
	return it.ForEach(handle)
}

// undeliveredTaskFetcher fetches tasks that were delivered to clients but never acknowledged.
type deliveredTaskFetcher struct {
	Tasks           *storage.TaskStore
	LastDeliveredID string
	Namespace       string
	AckDeadline     time.Duration
	watermark       time.Time
	lastDelivery    time.Time
}

func (dtf *deliveredTaskFetcher) Process(handle func(task *proto.Task) error) error {
	startTask, err := ksuid.Parse(dtf.LastDeliveredID)
	if err != nil {
		return err
	}
	dtf.lastDelivery = startTask.Time()

	dtf.watermark = time.Now().Add(-dtf.AckDeadline)
	it := storage.NewTaskIterator(dtf.Tasks.Client, &storage.ScheduledRange{
		Namespace: dtf.Namespace,
		StartID:   storage.NewTaskID(ksuid.Nil.Time()),
		EndID:     storage.NewTaskID(dtf.watermark),
	})

	return it.ForEach(func(task *proto.Task) error {
		if dtf.retryTask(task) {
			if err := handle(task); err != nil {
				return err
			}
		}
		return nil
	})
}

func (dtf *deliveredTaskFetcher) retryTask(task *proto.Task) bool {
	ts, err := ksuid.Parse(task.GetId())
	if err != nil {
		telemetry.Logger.Error("failed to parse task id",
			zap.Error(err))
		return false
	}

	lastRetry := task.GetLastRetry().AsTime()
	return !ts.Time().After(dtf.lastDelivery) &&
		(lastRetry.IsZero() || lastRetry.Add(dtf.AckDeadline).Before(dtf.watermark))
}
