package sched

import (
	"github.com/mlposey/z4/iden"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"io"
	"time"
)

// undeliveredTaskFetcher fetches tasks that have not been delivered to clients.
type undeliveredTaskFetcher struct {
	Tasks     *storage.TaskStore
	StartID   iden.TaskID
	Namespace string
	Prefetch  int
}

func (utf *undeliveredTaskFetcher) Process(handle func(task *proto.Task) error) error {
	it := storage.NewTaskIterator(utf.Tasks.Client, &storage.ScheduledRange{
		Namespace: utf.Namespace,
		StartID:   utf.StartID,
		EndID:     iden.New(time.Now(), 0),
		Prefetch:  utf.Prefetch,
	})

	first, err := it.Peek()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	if first.GetId() == utf.StartID.String() {
		// Skip last delivered tasks because we already delivered it :)
		it.Next()
	}
	return it.ForEach(handle)
}

// undeliveredTaskFetcher fetches tasks that were delivered to clients but never acknowledged.
type deliveredTaskFetcher struct {
	Tasks           *storage.TaskStore
	LastDeliveredID iden.TaskID
	Namespace       string
	AckDeadline     time.Duration
	watermark       time.Time
	lastDelivery    time.Time
}

func (dtf *deliveredTaskFetcher) Process(handle func(task *proto.Task) error) error {
	dtf.lastDelivery = dtf.LastDeliveredID.MustTime()

	dtf.watermark = time.Now().Add(-dtf.AckDeadline)
	it := storage.NewTaskIterator(dtf.Tasks.Client, &storage.ScheduledRange{
		Namespace: dtf.Namespace,
		StartID:   iden.Min,
		EndID:     iden.New(dtf.watermark, 0),
		Prefetch:  1_000,
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
	scheduleTime := iden.MustParseString(task.GetId()).MustTime()
	lastRetry := task.GetLastRetry().AsTime()

	return !scheduleTime.After(dtf.lastDelivery) &&
		(lastRetry.IsZero() || lastRetry.Add(dtf.AckDeadline).Before(dtf.watermark))
}
