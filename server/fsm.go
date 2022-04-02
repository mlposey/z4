package server

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"io"
)

type stateMachine struct {
	db *badger.DB
}

func newFSM(db *badger.DB) *stateMachine {
	return &stateMachine{db: db}
}

func (f *stateMachine) Apply(log *raft.Log) interface{} {
	panic("implement me")
}

func (f *stateMachine) Snapshot() (raft.FSMSnapshot, error) {
	return &snapshot{db: f.db}, nil
}

func (f *stateMachine) Restore(snapshot io.ReadCloser) error {
	// TODO: Freeze db writes
	return f.db.Load(snapshot, 100)
	// TODO: Resume db writes.
}

type snapshot struct {
	db *badger.DB
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	_, err := s.db.Backup(sink, 0)
	return err
}

func (s *snapshot) Release() {
	panic("implement me")
}
