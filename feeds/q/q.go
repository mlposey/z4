package q

import (
	"github.com/mlposey/z4/proto"
)

type TaskWriter interface {
	Push(tasks []*proto.Task) error
	Acknowledge(acks []*proto.Ack) error
}

type TaskReader interface {
	Tasks() TaskStream
}

type TaskStream <-chan *proto.Task
