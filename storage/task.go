package storage

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/iden"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	pb "google.golang.org/protobuf/proto"
)

// TaskStore manages persistent storage for tasks.
type TaskStore struct {
	Client *BadgerClient
}

func NewTaskStore(db *BadgerClient) *TaskStore {
	store := &TaskStore{Client: db}
	return store
}

func (ts *TaskStore) PurgeTasks(namespace string) error {
	fifoPrefix := getFifoPrefix(namespace)
	schedPrefix := getSchedPrefix(namespace)
	return ts.Client.DB.DropPrefix(fifoPrefix, schedPrefix)
}

func getFifoPrefix(namespace string) []byte {
	return []byte(fmt.Sprintf("task#fifo#%s#", namespace))
}

func getSchedPrefix(namespace string) []byte {
	return []byte(fmt.Sprintf("task#sched#%s#", namespace))
}

func (ts *TaskStore) DeleteAll(acks []*proto.Ack) error {
	telemetry.Logger.Debug("deleting task batch from DB", zap.Int("count", len(acks)))
	batch := ts.Client.DB.NewWriteBatch()
	defer batch.Cancel()

	for _, ack := range acks {
		key, err := getAckKey(ack)
		if err != nil {
			return fmt.Errorf("failed to delete task '%s' in batch: %w", ack.GetReference(), err)
		}

		err = batch.Delete(key)
		if err != nil {
			return fmt.Errorf("failed to delete task '%s' in batch: %w", ack.GetReference(), err)
		}
	}
	return batch.Flush()
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
		key := getTaskKey(task)
		err = batch.Set(key, payload)
		if err != nil {
			return fmt.Errorf("failed to write task '%s' from batch: %w", task.GetId(), err)
		}
	}
	return batch.Flush()
}

func (ts *TaskStore) Get(namespace string, id iden.TaskID) (*proto.Task, error) {
	var task *proto.Task
	err := ts.Client.DB.View(func(txn *badger.Txn) error {
		key := getScheduledTaskKey(namespace, id)
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
	return NewTaskIterator(ts.Client, query), nil
}

func getTaskKey(task *proto.Task) []byte {
	id := iden.MustParseString(task.GetId())
	if task.GetScheduleTime() == nil {
		return getFifoTaskKey(task.GetNamespace(), id)
	}
	return getScheduledTaskKey(task.GetNamespace(), id)
}

func getAckKey(ack *proto.Ack) ([]byte, error) {
	id, err := iden.ParseString(ack.GetReference().GetTaskId())
	if err != nil {
		return nil, err
	}

	_, err = id.Time()
	if errors.Is(err, iden.ErrNoTime) {
		return getFifoTaskKey(ack.GetReference().GetNamespace(), id), nil
	} else {
		return getScheduledTaskKey(ack.GetReference().GetNamespace(), id), nil
	}
}

func getScheduledTaskKey(namespaceID string, id iden.TaskID) []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(getSchedPrefix(namespaceID))
	buf.Write(id[:])
	return buf.Bytes()
}

func getFifoTaskKey(namespaceID string, id iden.TaskID) []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(getFifoPrefix(namespaceID))

	indexB := make([]byte, 8)
	binary.BigEndian.PutUint64(indexB, id.Index())
	buf.Write(indexB)

	return buf.Bytes()
}
