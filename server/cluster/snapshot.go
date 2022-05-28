package cluster

import (
	"bytes"
	"encoding/binary"
	"github.com/cockroachdb/pebble"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/telemetry"
	"io"
)

// snapshot is a point in time snapshot of the database.
//
// "Point in time" indicates that it is a reference to
// a time step where all data before that time should
// be saved.
type snapshot struct {
	db *pebble.Snapshot
}

func newSnapshot(db *pebble.DB) *snapshot {
	return &snapshot{db: db.NewSnapshot()}
}

// Persist saves all data from the snapshot to sink.
func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	telemetry.Logger.Info("persisting fsm snapshot...")
	defer telemetry.Logger.Info("snapshot persisted")

	it := s.db.NewIter(&pebble.IterOptions{})
	it.First()
	l := make([]byte, 4)
	for ; it.Valid(); it.Next() {
		binary.BigEndian.PutUint32(l, uint32(len(it.Key())))
		_, err := sink.Write(l)
		if err != nil {
			return err
		}

		_, err = sink.Write(it.Key())
		if err != nil {
			return err
		}

		binary.BigEndian.PutUint32(l, uint32(len(it.Value())))
		_, err = sink.Write(l)
		if err != nil {
			return err
		}

		_, err = sink.Write(it.Value())
		if err != nil {
			return err
		}
	}
	telemetry.LastFSMSnapshot.SetToCurrentTime()
	return it.Close()
}

func (s *snapshot) Release() {
	_ = s.db.Close()
}

// snapshotRestorer restores a persisted snapshot to the database.
type snapshotRestorer struct {
	db *pebble.DB
}

func newSnapshotRestorer(db *pebble.DB) *snapshotRestorer {
	return &snapshotRestorer{db: db}
}

// Restore loads all data from the snapshot into the database.
//
// Any existing data in the database will be removed prior to
// loading the snapshot.
//
// This method assumes the snapshot was created using the
// snapshot.Persist method.
func (s *snapshotRestorer) Restore(snapshot io.ReadCloser) error {
	err := s.clearDatabase()
	if err != nil {
		return err
	}
	return s.loadData(snapshot)
}

func (s *snapshotRestorer) clearDatabase() error {
	err := s.db.DeleteRange([]byte("task# "), []byte("task#z"), pebble.NoSync)
	if err != nil {
		return err
	}
	return s.db.DeleteRange([]byte("queue# "), []byte("queue#z"), pebble.NoSync)
}

func (s *snapshotRestorer) loadData(snapshot io.ReadCloser) error {
	l := make([]byte, 4)
	key := bytes.NewBuffer(nil)
	value := bytes.NewBuffer(nil)
	for {
		_, err := snapshot.Read(l)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		klen := int(binary.BigEndian.Uint32(l))
		if key.Cap() < klen {
			key.Grow(klen - key.Cap())
		}
		k := key.Next(klen)
		_, err = snapshot.Read(k)
		if err != nil {
			return err
		}

		_, err = snapshot.Read(l)
		if err != nil {
			return err
		}

		vlen := int(binary.BigEndian.Uint32(l))
		if value.Cap() < vlen {
			value.Grow(vlen - value.Cap())
		}
		v := key.Next(vlen)
		_, err = snapshot.Read(v)
		if err != nil {
			return err
		}

		err = s.db.Set(k, v, pebble.NoSync)
		if err != nil {
			return err
		}

		key.Reset()
		value.Reset()
	}
	return nil
}
