package q

import (
	"github.com/mlposey/z4/iden"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
)

type fifoDeliveredFactory struct {
}

func (f *fifoDeliveredFactory) Query(namespace *proto.Namespace) storage.TaskRange {
	lastID := namespace.LastDeliveredQueuedTask
	lastIndex := iden.MustParseString(lastID).Index()
	return &storage.FifoRange{
		Namespace:  namespace.GetId(),
		StartIndex: 0,
		EndIndex:   lastIndex,
		Prefetch:   1_000,
	}
}

type fifoUndeliveredFactory struct {
	prefetch int
}

func (f *fifoUndeliveredFactory) Query(namespace *proto.Namespace) storage.TaskRange {
	lastID := namespace.LastDeliveredQueuedTask
	lastIndex := iden.MustParseString(lastID).Index()
	nextIndex := lastIndex + 1

	return &storage.FifoRange{
		Namespace:  namespace.GetId(),
		StartIndex: nextIndex,
		// TODO: Make read limit configurable.
		EndIndex: nextIndex + 2_000, // must be greater than 1000 because the sequence buffer is 1000
		Prefetch: f.prefetch,
	}
}

func (f *fifoUndeliveredFactory) Inform(n int) {
	if n > 0 {
		f.prefetch = 1000
	} else {
		f.prefetch = 0
	}
}

type fifoCheckpointer struct {
}

func (f *fifoCheckpointer) Set(namespace *proto.Namespace, task *proto.Task) {
	namespace.LastDeliveredQueuedTask = task.GetId()
}
