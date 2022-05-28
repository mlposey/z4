package q

import (
	"github.com/mlposey/z4/iden"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"time"
)

type schedDeliveredQueryFactory struct {
}

func (s *schedDeliveredQueryFactory) Query(settings *proto.QueueConfig) storage.TaskRange {
	ackDeadline := time.Second * time.Duration(settings.GetAckDeadlineSeconds())
	watermark := time.Now().Add(-ackDeadline)
	return &storage.ScheduledRange{
		Queue:   settings.GetId(),
		StartID: iden.Min,
		EndID:   iden.New(watermark, 0),
	}
}

type schedUndeliveredQueryFactory struct {
}

func (s *schedUndeliveredQueryFactory) Query(settings *proto.QueueConfig) storage.TaskRange {
	lastID := iden.MustParseString(settings.LastDeliveredScheduledTask)
	nextID := iden.New(lastID.MustTime(), lastID.Index()+1)
	return &storage.ScheduledRange{
		Queue:   settings.GetId(),
		StartID: nextID,
		EndID:   iden.New(time.Now().UTC(), 0),
	}
}

type schedCheckpointer struct {
}

func (s *schedCheckpointer) Set(settings *proto.QueueConfig, task *proto.Task) {
	settings.LastDeliveredScheduledTask = task.GetId()
}
