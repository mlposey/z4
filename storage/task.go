package storage

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/telemetry"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	pb "google.golang.org/protobuf/proto"
	"io"
	"time"
)

// NewTaskID creates a task id based on the time it should be delivered.
//
// Task IDs are random strings can be lexicographically sorted according
// to the delivery time of the task.
func NewTaskID(deliverAt time.Time) string {
	id, err := ksuid.NewRandomWithTime(deliverAt)
	if err != nil {
		telemetry.Logger.Fatal("could not create task id", zap.Error(err))
	}
	return id.String()
}

// TaskStore manages persistent storage for tasks.
type TaskStore struct {
	Client *BadgerClient
}

func NewTaskStore(db *BadgerClient) *TaskStore {
	store := &TaskStore{Client: db}
	return store
}

func (ts *TaskStore) DeleteAll(acks []*proto.Ack) error {
	telemetry.Logger.Debug("deleting task batch from DB", zap.Int("count", len(acks)))
	batch := ts.Client.DB.NewWriteBatch()
	defer batch.Cancel()

	for _, ack := range acks {
		key := getTaskKey(ack.GetNamespace(), ack.GetTaskId())
		err := batch.Delete(key)
		if err != nil {
			return fmt.Errorf("failed to delete task '%s' in batch: %w", ack.GetTaskId(), err)
		}
	}
	return batch.Flush()
}

func (ts *TaskStore) Save(task *proto.Task) error {
	telemetry.Logger.Debug("writing task to DB", zap.Any("task", task))
	return ts.Client.DB.Update(func(txn *badger.Txn) error {
		payload, err := pb.Marshal(task)
		if err != nil {
			return fmt.Errorf("could not encode task: %w", err)
		}
		key := getTaskKey(task.GetNamespace(), task.GetId())
		return txn.Set(key, payload)
	})
}

func (ts *TaskStore) SaveAll(tasks []*proto.Task) error {
	telemetry.Logger.Debug("writing task batch to DB", zap.Int("count", len(tasks)))
	batch := ts.Client.DB.NewWriteBatch()
	defer batch.Cancel()

	for _, task := range tasks {
		payload, err := pb.Marshal(task)
		if err != nil {
			return fmt.Errorf("count not encode task '%s': %w", task.GetId(), err)
		}
		// TODO: Determine if grouping tasks by namespace before writing is beneficial.
		key := getTaskKey(task.GetNamespace(), task.GetId())
		err = batch.Set(key, payload)
		if err != nil {
			return fmt.Errorf("failed to write task '%s' from batch: %w", task.GetId(), err)
		}
	}
	return batch.Flush()
}

func (ts *TaskStore) Get(namespace, id string) (*proto.Task, error) {
	var task *proto.Task
	err := ts.Client.DB.View(func(txn *badger.Txn) error {
		key := getTaskKey(namespace, id)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			task = new(proto.Task)
			return pb.Unmarshal(val, task)
		})
	})
	return task, err
}

func (ts *TaskStore) IterateRange(query TaskRange) (*TaskIterator, error) {
	err := query.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks due to invalid query: %w", err)
	}
	return NewTaskIterator(ts.Client, query), nil
}

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

func getTaskKey(namespaceID, taskID string) []byte {
	return []byte(fmt.Sprintf("task#%s#%s", namespaceID, taskID))
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
