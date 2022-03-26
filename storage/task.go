package storage

import (
	"github.com/segmentio/ksuid"
	"time"
)

type Task struct {
	ID string
	// TODO: Consider putting namespace in here or the definition.
	TaskDefinition
}

func NewTask(def TaskDefinition) Task {
	return Task{
		ID:             ksuid.New().String(),
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
func (tr TaskRange) MinID() string {
	return tr.idFromTime(tr.Min)
}

// MaxID returns the id that would theoretically be at the end of the range.
func (tr TaskRange) MaxID() string {
	return tr.idFromTime(tr.Max)
}

func (tr TaskRange) idFromTime(ts time.Time) string {
	id, err := ksuid.NewRandomWithTime(tr.Min)
	if err != nil {
		panic(err)
	}
	return id.String()
}
