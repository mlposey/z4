package storage

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cockroachdb/pebble"
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
	batch := ts.Client.DB2.NewBatch()

	for _, ack := range acks {
		key, err := getAckKey(ack)
		if err != nil {
			telemetry.Logger.Error("invalid ack key",
				zap.Error(err), zap.String("namespace", ack.GetReference().GetNamespace()),
				zap.String("task_id", ack.GetReference().GetTaskId()))
			continue
		}

		err = batch.Delete(key, pebble.Sync)
		if err != nil {
			_ = batch.Close()
			return fmt.Errorf("failed to delete task '%s' in batch: %w", ack.GetReference(), err)
		}
	}
	return batch.Commit(pebble.Sync)
}

func (ts *TaskStore) SaveAll(tasks []*proto.Task, saveIndex bool) error {
	telemetry.Logger.Debug("writing task batch to DB", zap.Int("count", len(tasks)))
	batch := ts.Client.DB2.NewBatch()

	maxIndexes := make(map[string]uint64)
	for _, task := range tasks {
		id := iden.MustParseString(task.GetId())
		if saveIndex && id.Index() > maxIndexes[task.GetNamespace()] {
			maxIndexes[task.GetNamespace()] = id.Index()
		}

		payload, err := pb.Marshal(task)
		if err != nil {
			_ = batch.Close()
			return fmt.Errorf("count not encode task '%s': %w", task.GetId(), err)
		}

		key := getTaskKey(task, id)
		err = batch.Set(key, payload, pebble.Sync)
		if err != nil {
			_ = batch.Close()
			return fmt.Errorf("failed to write task '%s' from batch: %w", task.GetId(), err)
		}
	}

	// Save max indexes on followers in case they need to become a leader
	// and pick up the count.
	if saveIndex {
		for ns, index := range maxIndexes {
			var payload [8]byte
			binary.BigEndian.PutUint64(payload[:], index)
			err := batch.Set(getSeqKey(ns), payload[:], pebble.Sync)
			if err != nil {
				_ = batch.Close()
				return fmt.Errorf("failed to save index for namespace %s: %w", ns, err)
			}
		}
	}
	return batch.Commit(pebble.Sync)
}

func (ts *TaskStore) Get(namespace string, id iden.TaskID) (*proto.Task, error) {
	var task *proto.Task
	key := getScheduledTaskKey(namespace, id)
	item, closer, err := ts.Client.DB2.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	task = new(proto.Task)
	err = pb.Unmarshal(item, task)
	return task, err
}

func (ts *TaskStore) IterateRange(query TaskRange) (*TaskIterator, error) {
	return NewTaskIterator(ts.Client, query), nil
}

func getTaskKey(task *proto.Task, parsedID iden.TaskID) []byte {
	if task.GetScheduleTime() == nil {
		return getFifoTaskKey(task.GetNamespace(), parsedID)
	}
	return getScheduledTaskKey(task.GetNamespace(), parsedID)
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
