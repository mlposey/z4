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
	Client *PebbleClient
}

func NewTaskStore(db *PebbleClient) *TaskStore {
	store := &TaskStore{Client: db}
	return store
}

func getFifoPrefix(queue string) []byte {
	return []byte(fmt.Sprintf("task#fifo#%s#", queue))
}

func getSchedPrefix(queue string) []byte {
	return []byte(fmt.Sprintf("task#sched#%s#", queue))
}

func (ts *TaskStore) DeleteAll(acks []*proto.Ack) error {
	telemetry.Logger.Debug("deleting task batch from DB", zap.Int("count", len(acks)))
	batch := ts.Client.DB.NewBatch()

	for _, ack := range acks {
		key, err := getAckKey(ack)
		if err != nil {
			telemetry.Logger.Error("invalid ack key",
				zap.Error(err), zap.String("queue", ack.GetReference().GetQueue()),
				zap.String("task_id", ack.GetReference().GetTaskId()))
			continue
		}

		err = batch.Delete(key, pebble.NoSync)
		if err != nil {
			_ = batch.Close()
			return fmt.Errorf("failed to delete task '%s' in batch: %w", ack.GetReference(), err)
		}
	}
	return batch.Commit(pebble.NoSync)
}

func (ts *TaskStore) SaveAll(tasks []*proto.Task, saveIndex bool) error {
	telemetry.Logger.Debug("writing task batch to DB", zap.Int("count", len(tasks)))
	batch := ts.Client.DB.NewBatch()

	maxIndexes := make(map[string]uint64)
	for _, task := range tasks {
		id := iden.MustParseString(task.GetId())
		if saveIndex && id.Index() > maxIndexes[task.GetQueue()] {
			maxIndexes[task.GetQueue()] = id.Index()
		}

		payload, err := pb.Marshal(task)
		if err != nil {
			_ = batch.Close()
			return fmt.Errorf("count not encode task '%s': %w", task.GetId(), err)
		}

		key := getTaskKey(task, id)
		err = batch.Set(key, payload, pebble.NoSync)
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
			err := batch.Set(getSeqKey(ns), payload[:], pebble.NoSync)
			if err != nil {
				_ = batch.Close()
				return fmt.Errorf("failed to save index for queue %s: %w", ns, err)
			}
		}
	}
	return batch.Commit(pebble.NoSync)
}

func (ts *TaskStore) Get(queue string, id iden.TaskID) (*proto.Task, error) {
	var task *proto.Task
	key := getScheduledTaskKey(queue, id)
	item, closer, err := ts.Client.DB.Get(key)
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
		return getFifoTaskKey(task.GetQueue(), parsedID)
	}
	return getScheduledTaskKey(task.GetQueue(), parsedID)
}

func getAckKey(ack *proto.Ack) ([]byte, error) {
	id, err := iden.ParseString(ack.GetReference().GetTaskId())
	if err != nil {
		return nil, err
	}

	_, err = id.Time()
	if errors.Is(err, iden.ErrNoTime) {
		return getFifoTaskKey(ack.GetReference().GetQueue(), id), nil
	} else {
		return getScheduledTaskKey(ack.GetReference().GetQueue(), id), nil
	}
}

func getScheduledTaskKey(queue string, id iden.TaskID) []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(getSchedPrefix(queue))
	buf.Write(id[:])
	return buf.Bytes()
}

func getFifoTaskKey(queue string, id iden.TaskID) []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(getFifoPrefix(queue))

	indexB := make([]byte, 8)
	binary.BigEndian.PutUint64(indexB, id.Index())
	buf.Write(indexB)

	return buf.Bytes()
}
