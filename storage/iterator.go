package storage

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/proto"
	pb "google.golang.org/protobuf/proto"
	"io"
)

// TaskIterator iterates over a range of tasks in the database.
type TaskIterator struct {
	txn    *badger.Txn
	it     *badger.Iterator
	end    []byte
	prefix []byte
}

func NewTaskIterator(client *BadgerClient, query TaskRange) *TaskIterator {
	// TODO: Consider creating another type to pass in instead of BadgerClient.
	// We should try to avoid usage of BadgerClient in other packages as much as possible.

	txn := client.DB.NewTransaction(false)
	it := txn.NewIterator(badger.DefaultIteratorOptions)
	start := getTaskKey(query.Namespace, query.StartID)
	it.Seek(start)

	return &TaskIterator{
		txn:    txn,
		it:     it,
		end:    getTaskKey(query.Namespace, query.EndID),
		prefix: []byte(fmt.Sprintf("task#%s#", query.Namespace)),
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
	if !ti.it.ValidForPrefix(ti.prefix) {
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
	if !skipCheck && !ti.it.ValidForPrefix(ti.prefix) {
		return nil, io.EOF
	}

	item := ti.it.Item()
	if bytes.Compare(item.Key(), ti.end) > 0 {
		return nil, io.EOF
	}

	task := new(proto.Task)
	err := item.Value(func(val []byte) error {
		err := pb.Unmarshal(val, task)
		if err != nil {
			return err
		}

		if task.GetId() == "" {
			return errors.New("invalid task: empty id")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (ti *TaskIterator) Close() error {
	ti.it.Close()
	ti.txn.Discard()
	return nil
}

// TaskRange is a query for tasks within a time range.
type TaskRange struct {
	// Namespace restricts the search to only tasks in a given namespace.
	Namespace string

	// StartID restricts the search to all task IDs that are equal to it
	// or occur after it in ascending sorted order.
	StartID string

	// EndID restricts the search to all task IDs that are equal to it
	// or occur before it in ascending sorted order.
	EndID string
}

// Validate determines whether the TaskRange contains valid properties.
func (tr TaskRange) Validate() error {
	if tr.StartID == "" {
		return errors.New("missing start id in task range")
	}
	if tr.EndID == "" {
		return errors.New("missing end id in task range")
	}
	return nil
}
