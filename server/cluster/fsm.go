package cluster

import (
	"errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
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

	var tasks []*proto.Task
	var acks []*proto.Ack
	res := make([]interface{}, len(logs))

	for i, log := range logs {
		cmd := new(proto.Command)
		err := pb.Unmarshal(log.Data, cmd)
		if err != nil {
			res[i] = err
			return res
		}

		switch v := cmd.GetCmd().(type) {
		case *proto.Command_Task:
			tasks = append(tasks, v.Task)

		case *proto.Command_Ack:
			acks = append(acks, v.Ack)

		default:
			res[i] = errors.New("unknown command type: expected ack or task")
			return res
		}
	}

	// TODO: Determine if concurrently saving and deleting improves performance.

	if len(tasks) > 0 {
		err := f.ts.SaveAll(tasks)
		if err != nil {
			for i := 0; i < len(res); i++ {
				res[i] = err
			}
			return res
		}
	}

	if len(acks) > 0 {
		err := f.ts.DeleteAll(acks)
		if err != nil {
			for i := 0; i < len(res); i++ {
				res[i] = err
			}
			return res
		}
	}

	telemetry.AppliedLogs.Add(float64(len(acks) + len(tasks)))
	return res
}

func (f *stateMachine) Apply(log *raft.Log) interface{} {
	// Internally, this method should never be invoked. The raft package should
	// use our ApplyBatch method instead in order to speed up write performance.
	return errors.New("unexpected call to Apply")
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
