package raftutil_test

import (
	"github.com/google/uuid"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/storage/raftutil"
	"testing"
	"time"
)

func newPebbleClient(t *testing.T) *storage.PebbleClient {
	dir := "/tmp/" + uuid.New().String()
	db, err := storage.NewPebbleClient(dir)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestPebbleLogStore_FirstIndex(t *testing.T) {
	db := newPebbleClient(t)
	defer db.Close()

	var store raft.LogStore
	store, err := raftutil.NewLogStore(db.DB)
	if err != nil {
		t.Fatal(err)
	}

	start, end := 1, 100
	var logs []*raft.Log
	for i := start; i <= end; i++ {
		logs = append(logs, &raft.Log{
			Index:      uint64(i),
			Term:       1,
			Type:       raft.LogCommand,
			Data:       []byte("test"),
			AppendedAt: time.Now(),
		})
	}

	err = store.StoreLogs(logs)
	if err != nil {
		t.Fatal(err)
	}

	idx, err := store.FirstIndex()
	if idx != uint64(start) {
		t.Fatalf("wrong idx; got=%d expected%d", idx, start)
	}

	log := new(raft.Log)
	err = store.GetLog(idx, log)
	if err != nil {
		t.Fatal(err)
	}

	if log.Index != uint64(start) {
		t.Fatalf("wrong idx; got=%d expected%d", log.Index, start)
	}

	err = store.DeleteRange(uint64(start), 50)
	if err != nil {
		t.Fatal(err)
	}

	start = 51
	idx, err = store.FirstIndex()
	if idx != uint64(start) {
		t.Fatalf("wrong idx; got=%d expected%d", idx, start)
	}

	log = new(raft.Log)
	err = store.GetLog(idx, log)
	if err != nil {
		t.Fatal(err)
	}

	if log.Index != uint64(start) {
		t.Fatalf("wrong idx; got=%d expected%d", log.Index, start)
	}
}
