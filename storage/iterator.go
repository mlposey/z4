package storage

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/iden"
	"github.com/mlposey/z4/proto"
	pb "google.golang.org/protobuf/proto"
	"io"
	"time"
)

// TaskIterator iterates over a range of tasks in the database.
type TaskIterator struct {
	txn *badger.Txn
	it  *badger.Iterator
	end []byte
	buf []byte
}

func NewTaskIterator(client *BadgerClient, query TaskRange) *TaskIterator {
	// TODO: Consider creating another type to pass in instead of BadgerClient.
	// We should try to avoid usage of BadgerClient in other packages as much as possible.

	txn := client.DB.NewTransaction(false)
	opts := badger.DefaultIteratorOptions
	opts.Prefix = query.GetPrefix()
	if query.GetPrefetch() != 0 {
		opts.PrefetchSize = query.GetPrefetch()
	} else {
		opts.PrefetchValues = false
	}
	it := txn.NewIterator(opts)
	start := query.GetStart()
	it.Seek(start)

	return &TaskIterator{
		txn: txn,
		it:  it,
		end: query.GetEnd(),
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
	if !ti.it.Valid() {
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
	if !skipCheck && !ti.it.Valid() {
		return nil, io.EOF
	}

	item := ti.it.Item()
	if bytes.Compare(item.Key(), ti.end) > 0 {
		return nil, io.EOF
	}

	var err error
	ti.buf, err = item.ValueCopy(ti.buf)
	if err != nil {
		return nil, err
	}

	task := new(proto.Task)
	err = pb.Unmarshal(ti.buf, task)
	if err != nil {
		return nil, err
	}

	if task.GetId() == "" {
		return nil, errors.New("invalid task: empty id")
	}
	return task, nil
}

func (ti *TaskIterator) Close() error {
	ti.it.Close()
	ti.txn.Discard()
	return nil
}

type TaskRange interface {
	GetNamespace() string
	GetStart() []byte
	GetEnd() []byte
	GetPrefix() []byte
	GetPrefetch() int
}

// ScheduledRange is a query for tasks within a time range.
type ScheduledRange struct {
	// Namespace restricts the search to only tasks in a given namespace.
	Namespace string

	// StartID restricts the search to all task IDs that are equal to it
	// or occur after it in ascending sorted order.
	StartID iden.TaskID

	// EndID restricts the search to all task IDs that are equal to it
	// or occur before it in ascending sorted order.
	EndID iden.TaskID

	Prefetch int
}

func (tr *ScheduledRange) GetPrefix() []byte {
	return []byte(fmt.Sprintf("task#sched#%s#", tr.Namespace))
}

func (tr *ScheduledRange) GetNamespace() string {
	return tr.Namespace
}

func (tr *ScheduledRange) GetStart() []byte {
	return getScheduledTaskKey(tr.Namespace, tr.StartID)
}

func (tr *ScheduledRange) GetEnd() []byte {
	return getScheduledTaskKey(tr.Namespace, tr.EndID)
}

func (tr *ScheduledRange) GetPrefetch() int {
	return tr.Prefetch
}

// FifoRange is a query for tasks within an index range
type FifoRange struct {
	// Namespace restricts the search to only tasks in a given namespace.
	Namespace string

	StartIndex uint64
	EndIndex   uint64

	Prefetch int
}

func (fr *FifoRange) GetPrefix() []byte {
	return []byte(fmt.Sprintf("task#fifo#%s#", fr.Namespace))
}

func (fr *FifoRange) GetNamespace() string {
	return fr.Namespace
}

func (fr *FifoRange) GetStart() []byte {
	id := iden.New(time.Time{}, fr.StartIndex)
	return getFifoTaskKey(fr.Namespace, id)
}

func (fr *FifoRange) GetEnd() []byte {
	id := iden.New(time.Time{}, fr.EndIndex)
	return getFifoTaskKey(fr.Namespace, id)
}

func (fr *FifoRange) GetPrefetch() int {
	return fr.Prefetch
}
