package sm

import (
	"encoding/binary"
	"errors"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/telemetry"
	"io"
)

// SnapshotWriter creates snapshots of a database.
type SnapshotWriter struct {
	it KVIterator
}

func NewSnapshotWriter(it KVIterator) *SnapshotWriter {
	return &SnapshotWriter{it: it}
}

// Persist saves all data from the snapshot to the sink.
func (s *SnapshotWriter) Persist(sink raft.SnapshotSink) error {
	telemetry.Logger.Info("persisting fsm snapshot...")

	lb := make([]byte, 4)
	err := s.it.ForEach(func(k, v []byte) error {
		err := s.write(k, sink, lb)
		if err != nil {
			return err
		}
		return s.write(v, sink, lb)
	})
	if err != nil {
		return err
	}

	telemetry.Logger.Info("snapshot persisted")
	telemetry.LastFSMSnapshot.SetToCurrentTime()
	return nil
}

// writes a length-prefixed value to the destination
func (s *SnapshotWriter) write(src []byte, dst raft.SnapshotSink, lb []byte) error {
	err := checkLengthBuffer(lb)
	if err != nil {
		return err
	}

	// TODO: Consider using variable-length integers to store length.
	binary.BigEndian.PutUint32(lb, uint32(len(src)))
	_, err = dst.Write(lb)
	if err != nil {
		return err
	}

	_, err = dst.Write(src)
	return err
}

func (s *SnapshotWriter) Release() {
	_ = s.it.Close()
}

// SnapshotReader restores a persisted snapshot to the database.
type SnapshotReader struct {
	store KVStore
}

func NewSnapshotReader(store KVStore) *SnapshotReader {
	return &SnapshotReader{store: store}
}

// Restore loads all data from the snapshot into the database.
//
// Any existing data in the database will be removed prior to
// loading the snapshot.
//
// This method assumes the snapshot was created using the
// SnapshotWriter.Persist method.
func (s *SnapshotReader) Restore(snapshot io.ReadCloser) error {
	err := s.store.DeleteAll()
	if err != nil {
		return err
	}
	return s.loadData(snapshot)
}

func (s *SnapshotReader) loadData(snapshot io.ReadCloser) error {
	lb := make([]byte, 4)
	for {
		key, err := s.read(snapshot, lb)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		value, err := s.read(snapshot, lb)
		if err != nil {
			return err
		}

		err = s.store.Write(key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// reads a length-prefixed value from the source
func (s *SnapshotReader) read(src io.ReadCloser, lb []byte) ([]byte, error) {
	err := checkLengthBuffer(lb)
	if err != nil {
		return nil, err
	}

	_, err = io.ReadFull(src, lb)
	if err != nil {
		return nil, err
	}

	length := int(binary.BigEndian.Uint32(lb))
	val := make([]byte, length)
	_, err = io.ReadFull(src, val)
	return val, err
}

func checkLengthBuffer(lb []byte) error {
	if len(lb) != 4 {
		return errors.New("length buffer must be 4 bytes")
	}
	return nil
}
