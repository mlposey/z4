package sm

import (
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/feeds/q"
	"github.com/mlposey/z4/server/cluster/group"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"io"
)

// StateMachine uses raft logs to modify the task database.
type StateMachine struct {
	store    KVStore
	writer   *batchWriter
	restorer *SnapshotReader
}

func New(
	store KVStore,
	tw q.TaskWriter,
	ns *storage.SettingStore,
) *StateMachine {
	// TODO: Do not pass a *pebble.db directly. Create a new type.
	return &StateMachine{
		store:    store,
		writer:   newBatchWriter(tw, ns),
		restorer: NewSnapshotReader(store),
	}
}

func (f *StateMachine) SetHandle(handle *group.LeaderHandle) {
	f.writer.Handle = handle
}

func (f *StateMachine) ApplyBatch(logs []*raft.Log) []interface{} {
	telemetry.ReceivedLogs.Add(float64(len(logs)))
	return f.writer.Write(logs)
}

func (f *StateMachine) Apply(log *raft.Log) interface{} {
	return f.ApplyBatch([]*raft.Log{log})
}

func (f *StateMachine) Snapshot() (raft.FSMSnapshot, error) {
	telemetry.Logger.Info("taking fsm snapshot")
	return NewSnapshotWriter(f.store.Iterator()), nil
}

func (f *StateMachine) Restore(snapshot io.ReadCloser) error {
	telemetry.Logger.Info("restoring fsm from snapshot...")
	defer telemetry.Logger.Info("snapshot restored")

	return f.restorer.Restore(snapshot)
}
