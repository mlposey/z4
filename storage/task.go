package storage

import (
	"github.com/mlposey/z4/telemetry"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"time"
)

type Task struct {
	ID        string
	Namespace string
	RunTime   time.Time
	Metadata  map[string]string
	Payload   []byte
}

func NewTaskID(runTime time.Time) string {
	id, err := ksuid.NewRandomWithTime(runTime)
	if err != nil {
		telemetry.Logger.Fatal("could not create task id", zap.Error(err))
	}
	return id.String()
}

// TaskRange is a query for tasks within a time range.
type TaskRange struct {
	Namespace string
	Min       time.Time
	Max       time.Time
}

// MinID returns the id that would theoretically be at the start of the range.
func (tr TaskRange) MinID() string {
	return tr.idFromTime(tr.Min)
}

// MaxID returns the id that would theoretically be at the end of the range.
func (tr TaskRange) MaxID() string {
	return tr.idFromTime(tr.Max)
}

func (tr TaskRange) idFromTime(ts time.Time) string {
	id, err := ksuid.NewRandomWithTime(ts)
	if err != nil {
		panic(err)
	}
	return id.String()
}
