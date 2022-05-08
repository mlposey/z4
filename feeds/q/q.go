package q

import (
	"github.com/mlposey/z4/proto"
)

type TaskWriter interface {
	Push(task *proto.Task) error
	Acknowledge(taskID *proto.Task)
}

type TaskReader interface {
	Tasks() TaskStream
}

type TaskStream <-chan *proto.Task
