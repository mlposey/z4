package q

import (
	"github.com/mlposey/z4/proto"
)

type TaskWriter interface {
	Push(tasks []*proto.Task) error
	Acknowledge(acks []*proto.Ack) error
	NextIndex(namespace string) (uint64, error)
	PurgeTasks(namespace string) error
	Close() error
}

type TaskReader interface {
	Tasks() TaskStream
}

type TaskStream <-chan *proto.Task
