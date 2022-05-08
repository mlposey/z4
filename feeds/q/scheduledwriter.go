package q

import (
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
)

type scheduledTaskWriter struct {
	tasks *storage.TaskStore
}

func NewScheduledTaskWriter(tasks *storage.TaskStore) TaskWriter {
	return &scheduledTaskWriter{tasks: tasks}
}

func (s *scheduledTaskWriter) Push(tasks []*proto.Task) error {
	return s.tasks.SaveAll(tasks)
}

func (s *scheduledTaskWriter) Acknowledge(acks []*proto.Ack) error {
	return s.tasks.DeleteAll(acks)
}
