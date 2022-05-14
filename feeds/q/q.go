package q

import (
	"github.com/mlposey/z4/proto"
)

type TaskWriter interface {
	Push(tasks []*proto.Task, isFollower bool) error
	Acknowledge(acks []*proto.Ack) error
	PurgeTasks(namespace string) error
	Close() error
}

type TaskReader interface {
	Tasks() TaskStream
}

type TaskStream <-chan *proto.Task
