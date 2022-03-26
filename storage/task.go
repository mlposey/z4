package storage

import (
	"github.com/mlposey/z4/telemetry"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"time"
)

type Task struct {
	ID ksuid.KSUID
	// TODO: Consider putting namespace in here or the definition.
	TaskDefinition
}

func NewTask(def TaskDefinition) Task {
	id, err := ksuid.NewRandomWithTime(def.RunTime)
	if err != nil {
		telemetry.Logger.Fatal("could not create task id", zap.Error(err))
	}
	return Task{
		ID:             id,
		TaskDefinition: def,
	}
}

type TaskDefinition struct {
	RunTime  time.Time
	Metadata map[string]string
	Payload  []byte
}

// TaskRange is a query for tasks within a time range.
type TaskRange struct {
	Namespace string
	Min       time.Time
	Max       time.Time
}

// MinID returns the id that would theoretically be at the start of the range.
func (tr TaskRange) MinID() ksuid.KSUID {
	return tr.idFromTime(tr.Min)
}

// MaxID returns the id that would theoretically be at the end of the range.
func (tr TaskRange) MaxID() ksuid.KSUID {
	return tr.idFromTime(tr.Max)
}

func (tr TaskRange) idFromTime(ts time.Time) ksuid.KSUID {
	id, err := ksuid.NewRandomWithTime(ts)
	if err != nil {
		panic(err)
	}
	return id
}
