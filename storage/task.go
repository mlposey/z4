package storage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/telemetry"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	pb "google.golang.org/protobuf/proto"
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
		key := getAckKey(ack)
		err := batch.Delete(key)
		if err != nil {
			return fmt.Errorf("failed to delete task '%s' in batch: %w", ack.GetTaskId(), err)
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

func (ts *TaskStore) Get(namespace, id string) (*proto.Task, error) {
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
	err := query.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks due to invalid query: %w", err)
	}
	return NewTaskIterator(ts.Client, query), nil
}

func getTaskKey(task *proto.Task) []byte {
	if task.GetScheduleTime() == nil {
		return getFifoTaskKey(task.GetNamespace(), task.GetIndex())
	}
	return getScheduledTaskKey(task.GetNamespace(), task.GetId())
}

func getAckKey(ack *proto.Ack) []byte {
	if ack.GetTaskId() == "" {
		return getFifoTaskKey(ack.GetNamespace(), ack.GetIndex())
	}
	return getScheduledTaskKey(ack.GetNamespace(), ack.GetTaskId())
}

func getScheduledTaskKey(namespaceID, taskID string) []byte {
	return []byte(fmt.Sprintf("task#sched#%s#%s", namespaceID, taskID))
}

func getFifoTaskKey(namespaceID string, index uint64) []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf("task#fifo#%s#", namespaceID))

	/*indexB := make([]byte, 8)
	binary.BigEndian.PutUint64(indexB, index)
	buf.Write(indexB)*/
	_ = binary.Write(buf, binary.BigEndian, index)

	return buf.Bytes()
}
