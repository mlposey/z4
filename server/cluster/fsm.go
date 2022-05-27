package cluster

import (
	"github.com/cockroachdb/pebble"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/feeds/q"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"io"
)

// stateMachine uses raft logs to modify the task database.
type stateMachine struct {
	db       *pebble.DB
	writer   *batchWriter
	restorer *snapshotRestorer
}

func newFSM(
	db *pebble.DB,
	tw q.TaskWriter,
	ns *storage.NamespaceStore,
) *stateMachine {
	// TODO: Do not pass a *pebble.db directly. Create a new type.
	return &stateMachine{
		db:       db,
		writer:   newBatchWriter(tw, ns),
		restorer: newSnapshotRestorer(db),
	}
}

func (f *stateMachine) SetHandle(handle *LeaderHandle) {
	f.writer.Handle = handle
}

func (f *stateMachine) ApplyBatch(logs []*raft.Log) []interface{} {
	telemetry.ReceivedLogs.Add(float64(len(logs)))
	return f.writer.Write(logs)
}

func (f *stateMachine) Apply(log *raft.Log) interface{} {
	return f.ApplyBatch([]*raft.Log{log})
}

func (f *stateMachine) Snapshot() (raft.FSMSnapshot, error) {
	telemetry.Logger.Info("taking fsm snapshot")
	return newSnapshot(f.db), nil
}

func (f *stateMachine) Restore(snapshot io.ReadCloser) error {
	telemetry.Logger.Info("restoring fsm from snapshot...")
	defer telemetry.Logger.Info("snapshot restored")

	return f.restorer.Restore(snapshot)
}
