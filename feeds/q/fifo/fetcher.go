package fifo

import (
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"time"
)

// undeliveredTaskFetcher fetches tasks that have not been delivered to clients.
type undeliveredTaskFetcher struct {
	Tasks      *storage.TaskStore
	StartIndex uint64
	Namespace  string
	Prefetch   int
}

func (utf *undeliveredTaskFetcher) Process(handle func(task *proto.Task) error) error {
	it := storage.NewTaskIterator(utf.Tasks.Client, &storage.FifoRange{
		Namespace:  utf.Namespace,
		StartIndex: utf.StartIndex,
		// TODO: Make read limit configurable.
		EndIndex: utf.StartIndex + 2_000, // must be greater than 1000 because the sequence buffer is 1000
		Prefetch: utf.Prefetch,
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
		Prefetch:   1_000,
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
	if task.GetLastRetry() == nil {
		return task.GetCreatedAt().AsTime().Before(dtf.watermark)
	} else {
		return task.GetLastRetry().AsTime().Before(dtf.watermark)
	}
}
