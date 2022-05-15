package q

import (
	"github.com/mlposey/z4/iden"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"time"
)

type schedDeliveredQueryFactory struct {
}

func (s *schedDeliveredQueryFactory) Query(namespace *proto.Namespace) storage.TaskRange {
	ackDeadline := time.Second * time.Duration(namespace.GetAckDeadlineSeconds())
	watermark := time.Now().Add(-ackDeadline)
	return &storage.ScheduledRange{
		Namespace: namespace.GetId(),
		StartID:   iden.Min,
		EndID:     iden.New(watermark, 0),
		Prefetch:  1_000,
	}
}

type schedUndeliveredQueryFactory struct {
	prefetch int
}

func (s *schedUndeliveredQueryFactory) Query(namespace *proto.Namespace) storage.TaskRange {
	lastID := iden.MustParseString(namespace.LastDeliveredScheduledTask)
	nextID := iden.New(lastID.MustTime(), lastID.Index()+1)
	return &storage.ScheduledRange{
		Namespace: namespace.GetId(),
		StartID:   nextID,
		EndID:     iden.New(time.Now(), 0),
		Prefetch:  s.prefetch,
	}
}

func (s *schedUndeliveredQueryFactory) Inform(n int) {
	if n > 0 {
		s.prefetch = 1000
	} else {
		s.prefetch = 0
	}
}

type schedCheckpointer struct {
}

func (s *schedCheckpointer) Set(namespace *proto.Namespace, task *proto.Task) {
	namespace.LastDeliveredScheduledTask = task.GetId()
}
