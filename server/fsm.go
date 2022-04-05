package server

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/feeds"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	pb "google.golang.org/protobuf/proto"
	"io"
)

type stateMachine struct {
	db *badger.DB
	fm *feeds.Manager
}

func newFSM(db *badger.DB, fm *feeds.Manager) *stateMachine {
	return &stateMachine{
		db: db,
		fm: fm,
	}
}

func (f *stateMachine) Apply(log *raft.Log) interface{} {
	task := new(proto.Task)
	err := pb.Unmarshal(log.Data, task)
	if err != nil {
		telemetry.Logger.Error("failed to parse log data", zap.Error(err))
		return err
	}

	lease := f.fm.Lease(task.GetNamespace())
	defer lease.Release()

	// TODO: Consider async.
	// Not currently sure how that will impact logs during sudden exits
	// when the batches can't be properly flushed.
	// There is an Apply that provides a batch of logs. Maybe we can
	// skip async writes and simply write the batch of tasks to badger
	// in a synchronous way.
	err = lease.Feed().Add(task)
	if err != nil {
		telemetry.Logger.Error("task submission to feed failed", zap.Error(err))
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
		sink.Close()
	}
	return err
}

func (s *snapshot) Release() {
}
