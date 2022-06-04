package sm_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/mlposey/z4/server/cluster/sm"
	"github.com/mlposey/z4/telemetry"
	"os"
	"testing"
)

func init() {
	_ = telemetry.InitLogger(true)
}

// TODO: Test that the database is cleared before adding items from snapshot.

// TestSnapshot verifies that a database snapshot can be created and restored.
func TestSnapshot(t *testing.T) {
	items := getItems()

	// Create a database and add our items to it.
	database := newKvStore()
	for key, value := range items {
		k, _ := base64.StdEncoding.DecodeString(key)
		err := database.Write(k, value)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create a place to store our snapshot.
	snapshots := newSnapshotStore()

	// Persist all database items to the snapshot store.
	writer := sm.NewSnapshotWriter(database)
	err := writer.Persist(snapshots)
	if err != nil {
		t.Fatal(err)
	}
	writer.Release()

	// Create a new, empty database.
	err = database.DeleteAll()
	if err != nil {
		t.Fatal(err)
	}
	database = newKvStore()

	// Restore the snapshot to the new database.
	reader := sm.NewSnapshotReader(database)
	err = reader.Restore(snapshots)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that all items from the original set were
	// restored to this new database.
	var count int
	err = database.ForEach(func(k []byte, v []byte) error {
		count++
		val := items[base64.StdEncoding.EncodeToString(k)]
		if !bytes.Equal(val, v) {
			t.Fatal("snapshot was missing item")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if count != len(items) {
		t.Fatal("snapshot contained more items than expected")
	}
}

func getItems() map[string][]byte {
	items := make(map[string][]byte)
	for i := 0; i < os.Getpagesize()*2; i++ {
		key := []byte(fmt.Sprintf("key-%d", i))
		value := []byte(fmt.Sprintf("value-%d", i))
		items[base64.StdEncoding.EncodeToString(key)] = value
	}
	return items
}

type kvStore struct {
	data map[string][]byte
}

func newKvStore() *kvStore {
	return &kvStore{data: make(map[string][]byte)}
}

func (s *kvStore) ForEach(visit func(k []byte, v []byte) error) error {
	for k, v := range s.data {
		key, _ := base64.StdEncoding.DecodeString(k)
		err := visit(key, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *kvStore) Close() error {
	return s.DeleteAll()
}

func (s *kvStore) Write(k, v []byte) error {
	s.data[base64.StdEncoding.EncodeToString(k)] = v
	return nil
}

func (s *kvStore) DeleteAll() error {
	s.data = make(map[string][]byte)
	return nil
}

func (s *kvStore) Iterator() sm.KVIterator {
	return s
}

type snapshotStore struct {
	snapshot *bytes.Buffer
}

func newSnapshotStore() *snapshotStore {
	return &snapshotStore{snapshot: bytes.NewBuffer(nil)}
}

func (s *snapshotStore) Read(p []byte) (n int, err error) {
	return s.snapshot.Read(p)
}

func (s *snapshotStore) Write(p []byte) (n int, err error) {
	return s.snapshot.Write(p)
}

func (s *snapshotStore) Close() error {
	return nil
}

func (s *snapshotStore) ID() string {
	return "test"
}

func (s *snapshotStore) Cancel() error {
	return nil
}
