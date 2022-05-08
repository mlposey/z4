package q

import (
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
)

type taskWriter struct {
	tasks *storage.TaskStore
}

func NewTaskWriter(tasks *storage.TaskStore) TaskWriter {
	return &taskWriter{tasks: tasks}
}

func (s *taskWriter) Push(tasks []*proto.Task) error {
	return s.tasks.SaveAll(tasks)
}

func (s *taskWriter) Acknowledge(acks []*proto.Ack) error {
	return s.tasks.DeleteAll(acks)
}
