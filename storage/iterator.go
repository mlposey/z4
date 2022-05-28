package storage

import (
	"bytes"
	"errors"
	"github.com/cockroachdb/pebble"
	"github.com/mlposey/z4/iden"
	"github.com/mlposey/z4/proto"
	pb "google.golang.org/protobuf/proto"
	"io"
	"time"
)

// TaskIterator iterates over a range of tasks in the database.
type TaskIterator struct {
	it     *pebble.Iterator
	end    []byte
	prefix []byte
}

func NewTaskIterator(client *PebbleClient, query TaskRange) *TaskIterator {
	// TODO: Consider creating another type to pass in instead of PebbleClient.
	// We should try to avoid usage of PebbleClient in other packages as much as possible.

	it := client.DB.NewIter(&pebble.IterOptions{})
	start := query.GetStart()
	it.SeekGE(start)

	return &TaskIterator{
		it:     it,
		end:    query.GetEnd(),
		prefix: query.GetPrefix(),
	}
}

// ForEach calls handle on the result of Next until an error occurs or no tasks remain.
func (ti *TaskIterator) ForEach(handle func(task *proto.Task) error) error {
	for {
		task, err := ti.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		err = handle(task)
		if err != nil {
			return err
		}
	}
}

func (ti *TaskIterator) Next() (*proto.Task, error) {
	if !(ti.it.Valid() && bytes.HasPrefix(ti.it.Key(), ti.prefix)) {
		return nil, io.EOF
	}

	task, err := ti.peek(true)
	if err != io.EOF {
		ti.it.Next()
	}
	return task, err
}

func (ti *TaskIterator) Peek() (*proto.Task, error) {
	return ti.peek(false)
}

func (ti *TaskIterator) peek(skipCheck bool) (*proto.Task, error) {
	if !skipCheck && !(ti.it.Valid() && bytes.HasPrefix(ti.it.Key(), ti.prefix)) {
		return nil, io.EOF
	}

	if bytes.Compare(ti.it.Key(), ti.end) > 0 {
		return nil, io.EOF
	}

	task := new(proto.Task)
	err := pb.Unmarshal(ti.it.Value(), task)
	if err != nil {
		return nil, err
	}

	if task.GetId() == "" {
		return nil, errors.New("invalid task: empty id")
	}
	return task, nil
}

func (ti *TaskIterator) Close() error {
	return ti.it.Close()
}

type TaskRange interface {
	GetQueue() string
	GetStart() []byte
	GetEnd() []byte
	GetPrefix() []byte
}

// ScheduledRange is a query for tasks within a time range.
type ScheduledRange struct {
	// Queue restricts the search to only tasks in a given queue.
	Queue string

	// StartID restricts the search to all task IDs that are equal to it
	// or occur after it in ascending sorted order.
	StartID iden.TaskID

	// EndID restricts the search to all task IDs that are equal to it
	// or occur before it in ascending sorted order.
	EndID iden.TaskID
}

func (tr *ScheduledRange) GetPrefix() []byte {
	return getSchedPrefix(tr.Queue)
}

func (tr *ScheduledRange) GetQueue() string {
	return tr.Queue
}

func (tr *ScheduledRange) GetStart() []byte {
	return getScheduledTaskKey(tr.Queue, tr.StartID)
}

func (tr *ScheduledRange) GetEnd() []byte {
	return getScheduledTaskKey(tr.Queue, tr.EndID)
}

// FifoRange is a query for tasks within an index range
type FifoRange struct {
	// Queue restricts the search to only tasks in a given queue.
	Queue string

	StartIndex uint64
	EndIndex   uint64
}

func (fr *FifoRange) GetPrefix() []byte {
	return getFifoPrefix(fr.Queue)
}

func (fr *FifoRange) GetQueue() string {
	return fr.Queue
}

func (fr *FifoRange) GetStart() []byte {
	id := iden.New(time.Time{}, fr.StartIndex)
	return getFifoTaskKey(fr.Queue, id)
}

func (fr *FifoRange) GetEnd() []byte {
	id := iden.New(time.Time{}, fr.EndIndex)
	return getFifoTaskKey(fr.Queue, id)
}
