package storage

import (
	"errors"
	"github.com/mlposey/z4/telemetry"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"time"
)

// NewTaskID creates a task id based on the time it should be delivered.
//
// Task IDs are random strings can be lexicographically sorted according
// to the delivery time of the task.
func NewTaskID(deliverAt time.Time) string {
	id, err := ksuid.NewRandomWithTime(deliverAt)
	if err != nil {
		telemetry.Logger.Fatal("could not create task id", zap.Error(err))
	}
	return id.String()
}

// TaskRange is a query for tasks within a time range.
type TaskRange struct {
	// Namespace restricts the search to only tasks in a given namespace.
	Namespace string

	// StartID restricts the search to all task IDs that are equal to it
	// or occur after it in ascending sorted order.
	StartID string

	// EndID restricts the search to all task IDs that are equal to it
	// or occur before it in ascending sorted order.
	EndID string
}

func (tr TaskRange) Validate() error {
	if tr.StartID == "" {
		return errors.New("missing start id in task range")
	}
	if tr.EndID == "" {
		return errors.New("missing end id in task range")
	}
	return nil
}
