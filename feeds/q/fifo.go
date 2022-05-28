package q

import (
	"github.com/mlposey/z4/iden"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"math"
)

type fifoDeliveredFactory struct {
}

func (f *fifoDeliveredFactory) Query(settings *proto.QueueConfig) storage.TaskRange {
	lastID := settings.LastDeliveredQueuedTask
	lastIndex := iden.MustParseString(lastID).Index()
	return &storage.FifoRange{
		Queue:      settings.GetId(),
		StartIndex: 0,
		EndIndex:   lastIndex,
	}
}

type fifoUndeliveredFactory struct {
}

func (f *fifoUndeliveredFactory) Query(settings *proto.QueueConfig) storage.TaskRange {
	lastID := settings.LastDeliveredQueuedTask
	lastIndex := iden.MustParseString(lastID).Index()
	nextIndex := lastIndex + 1

	return &storage.FifoRange{
		Queue:      settings.GetId(),
		StartIndex: nextIndex,
		EndIndex:   math.MaxUint64,
	}
}

type fifoCheckpointer struct {
}

func (f *fifoCheckpointer) Set(settings *proto.QueueConfig, task *proto.Task) {
	settings.LastDeliveredQueuedTask = task.GetId()
}
