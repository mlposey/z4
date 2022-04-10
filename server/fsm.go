package server

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	pb "google.golang.org/protobuf/proto"
	"io"
)

// stateMachine uses raft logs to modify the task database.
type stateMachine struct {
	db *badger.DB
	ts *storage.TaskStore
}

func newFSM(db *badger.DB, ts *storage.TaskStore) *stateMachine {
	return &stateMachine{
		db: db,
		ts: ts,
	}
}

func (f *stateMachine) ApplyBatch(logs []*raft.Log) []interface{} {
	telemetry.ReceivedLogs.Add(float64(len(logs)))
	tasks := make([]*proto.Task, len(logs))
	var taskCount float64
	res := make([]interface{}, len(logs))

	for i, log := range logs {
		task := new(proto.Task)
		err := pb.Unmarshal(log.Data, task)
		if err != nil {
			res[i] = err
			return res
		}

		tasks[i] = task
		taskCount++
	}

	err := f.ts.SaveAll(tasks)
	if err != nil {
		for i := 0; i < len(res); i++ {
			res[i] = err
		}
	} else {
		telemetry.AppliedLogs.Add(taskCount)
	}
	return res
}

func (f *stateMachine) Apply(log *raft.Log) interface{} {
	// Internally, this method should never be invoked. The raft package should
	// use our ApplyBatch method instead in order to speed up write performance.
	// However, we will keep this implementation around just in case the raft
	// package uses it in the future.

	telemetry.ReceivedLogs.Inc()
	task := new(proto.Task)
	err := pb.Unmarshal(log.Data, task)
	if err != nil {
		telemetry.Logger.Error("failed to parse log data", zap.Error(err))
		return err
	}

	err = f.ts.Save(task)
	if err != nil {
		telemetry.Logger.Error("task submission to feed failed", zap.Error(err))
	} else {
		telemetry.AppliedLogs.Inc()
	}
	return err
}

func (f *stateMachine) Snapshot() (raft.FSMSnapshot, error) {
	telemetry.Logger.Info("taking fsm snapshot")
	return &snapshot{db: f.db}, nil
}

func (f *stateMachine) Restore(snapshot io.ReadCloser) error {
	telemetry.Logger.Info("restoring fsm from snapshot")
	// TODO: Freeze db writes
	return f.db.Load(snapshot, 100)
	// TODO: Resume db writes.
}

type snapshot struct {
	db *badger.DB
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	telemetry.Logger.Info("persisting fsm snapshot")
	_, err := s.db.Backup(sink, 0)
	if err != nil {
		sink.Cancel()
	} else {
		telemetry.LastFSMSnapshot.SetToCurrentTime()
		sink.Close()
	}
	return err
}

func (s *snapshot) Release() {
}
