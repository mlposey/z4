package q

import (
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"time"
)

// A TaskWriter writes task-related entities to a collection.
type TaskWriter interface {
	// Push should add tasks to the collection.
	//
	// isFollower indicates whether this operation is
	// being performed on a cluster follower (as opposed
	// to the leader).
	Push(tasks []*proto.Task, isFollower bool) error

	// Acknowledge should apply task acknowledgements
	// to the collection.
	Acknowledge(acks []*proto.Ack) error

	// PurgeTasks should delete all tasks in the collection.
	PurgeTasks(namespace string) error

	// Close should release all resources needed to
	// safely destroy the TaskWriter object.
	Close() error
}

// A TaskStream is an unbounded collection of tasks.
type TaskStream <-chan *proto.Task

// A Reader reads tasks from a collection.
type Reader interface {
	// Ready signals whether Read *should* be invoked.
	//
	// This can be used to prevent concurrent Reads
	// if the Read operation will take a long time.
	Ready() bool

	// Read should call f on each task in the collection.
	Read(f func(task *proto.Task) error) error
}

// A QueryFactory generates task queries.
type QueryFactory interface {
	Query(namespace *proto.Namespace) storage.TaskRange
}

// A PreemptiveQueryFactory generates task queries.
//
// It is much like a QueryFactory but contains an
// additional method Inform that can help the factory
// build queries based on the results of previous queries.
type PreemptiveQueryFactory interface {
	QueryFactory
	Inform(n int)
}

// A Checkpointer saves a checkpoint based on a task.
type Checkpointer interface {
	Set(namespace *proto.Namespace, task *proto.Task)
}

// Readers instantiates a list of all possible readers.
func Readers(
	tasks *storage.TaskStore,
	namespace *proto.Namespace,
) []Reader {
	redeliveryInterval := time.Second * 30
	return []Reader{
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
