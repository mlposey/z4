package fifo

import (
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"time"
)

// undeliveredTaskFetcher fetches tasks that have not been delivered to clients.
type undeliveredTaskFetcher struct {
	Tasks      *storage.TaskStore
	StartIndex uint64
	Namespace  string
}

func (utf *undeliveredTaskFetcher) Process(handle func(task *proto.Task) error) error {
	it := storage.NewTaskIterator(utf.Tasks.Client, &storage.FifoRange{
		Namespace:  utf.Namespace,
		StartIndex: utf.StartIndex,
		// TODO: Make read limit configurable.
		EndIndex: utf.StartIndex + 1000,
	})
	return it.ForEach(handle)
}

// undeliveredTaskFetcher fetches tasks that were delivered to clients but never acknowledged.
type deliveredTaskFetcher struct {
	Tasks              *storage.TaskStore
	LastDeliveredIndex uint64
	Namespace          string
	AckDeadline        time.Duration
	watermark          time.Time
	lastDelivery       time.Time
}

func (dtf *deliveredTaskFetcher) Process(handle func(task *proto.Task) error) error {
	dtf.watermark = time.Now().Add(-dtf.AckDeadline)
	it := storage.NewTaskIterator(dtf.Tasks.Client, &storage.FifoRange{
		Namespace:  dtf.Namespace,
		StartIndex: 0,
		EndIndex:   dtf.LastDeliveredIndex,
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
	id, err := ksuid.Parse(task.GetId())
	if err != nil {
		telemetry.Logger.Error("could not parse id from fifo task",
			zap.Error(err))
		return false
	}
	createdAt := id.Time()

	if task.GetLastRetry() == nil {
		return createdAt.Before(dtf.watermark)
	} else {
		return task.GetLastRetry().AsTime().Before(dtf.watermark)
	}
}
