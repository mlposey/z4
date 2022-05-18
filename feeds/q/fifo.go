package q

import (
	"github.com/mlposey/z4/iden"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"math"
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
	}
}

type fifoUndeliveredFactory struct {
}

func (f *fifoUndeliveredFactory) Query(namespace *proto.Namespace) storage.TaskRange {
	lastID := namespace.LastDeliveredQueuedTask
	lastIndex := iden.MustParseString(lastID).Index()
	nextIndex := lastIndex + 1

	return &storage.FifoRange{
		Namespace:  namespace.GetId(),
		StartIndex: nextIndex,
		EndIndex:   math.MaxUint64,
	}
}

type fifoCheckpointer struct {
}

func (f *fifoCheckpointer) Set(namespace *proto.Namespace, task *proto.Task) {
	namespace.LastDeliveredQueuedTask = task.GetId()
}
