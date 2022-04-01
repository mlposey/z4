package storage

import (
	"github.com/mlposey/z4/telemetry"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"time"
)

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
	StartID   string
	EndID     string
}
