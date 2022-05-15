package q

import (
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"time"
)

type TaskWriter interface {
	Push(tasks []*proto.Task, isFollower bool) error
	Acknowledge(acks []*proto.Ack) error
	PurgeTasks(namespace string) error
	Close() error
}

type TaskStream <-chan *proto.Task

type ReadOperation interface {
	Ready() bool
	Run(f func(task *proto.Task) error) error
}

type QueryFactory interface {
	Query(namespace *proto.Namespace) storage.TaskRange
}

type PreemptiveQueryFactory interface {
	QueryFactory
	Inform(n int)
}

type Checkpointer interface {
	Set(namespace *proto.Namespace, task *proto.Task)
}

func ReadOperations(
	tasks *storage.TaskStore,
	namespace *proto.Namespace,
) []ReadOperation {
	redeliveryInterval := time.Second * 30
	return []ReadOperation{
		// Handle redelivery for queued tasks.
		newDeliveredReader(
			redeliveryInterval,
			tasks,
			namespace,
			new(fifoDeliveredFactory),
		),
		// Handle redelivery for scheduled tasks.
		newDeliveredReader(
			redeliveryInterval,
			tasks,
			namespace,
			new(schedDeliveredQueryFactory),
		),
		// Handle delivery for queued tasks.
		newUndeliveredReader(
			tasks,
			namespace,
			new(fifoUndeliveredFactory),
			new(fifoCheckpointer),
		),
		// Handle delivery for scheduled tasks.
		newUndeliveredReader(
			tasks,
			namespace,
			new(schedUndeliveredQueryFactory),
			new(schedCheckpointer),
		),
	}
}
