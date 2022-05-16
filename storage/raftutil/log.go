package raftutil

import (
	"bytes"
	"encoding/binary"
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"github.com/vmihailenco/msgpack/v5"
	"math"
)

type BadgerLogStore struct {
	db              *badger.DB
	cache           *logCache
	maxCompactedIdx uint64
}

var _ raft.LogStore = (*BadgerLogStore)(nil)
var logStorePrefix = []byte("raft#logstore#")

func NewLogStore(db *badger.DB) (*BadgerLogStore, error) {
	return &BadgerLogStore{
		db: db,
		// TODO: Find a good cache size.
		// This value seems to work well in practice, but it was
		// really just pulled from a hat.
		cache: newLogCache(db, 100_000),
	}, nil
}

func (b *BadgerLogStore) FirstIndex() (uint64, error) {
	var index uint64
	err := b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = logStorePrefix

		it := txn.NewIterator(opts)
		defer it.Close()
		it.Seek(getLogKey(0))
		if !it.Valid() {
			return nil
		}

		k := it.Item().Key()
		idx := k[len(k)-8:]
		index = binary.BigEndian.Uint64(idx)
		return nil
	})
	if err != nil {
		return 0, err
	}

	if index < b.maxCompactedIdx {
		// TODO: Properly fix FirstIndex compaction bug.
		// This is technically a fix, but it is more of a workaround.

		// This block handles a weird but reproducible bug that
		// I don't quite understand.
		//
		// When the Raft log becomes large enough, it is compacted.
		// Compaction involves a few steps
		// 1. Take a snapshot of the finite state machine (aka the
		//    task database).
		// 2. Select the first N logs for deletion.
		//    Example:
		//      N = 1000
		//      First log has index 50
		//      The following indexes are selected (inclusive): [50, 1050]
		// 3. Use the DeleteRange method of the LogStore to actually
		//    perform the log deletions. Once this is complete, the
		//    first log will change to be the last compacted log plus one.
		//
		// It has been observed that, after log compaction, the FirstIndex
		// method of the LogStore occasionally returns an index that has
		// recently been compacted. I.e., it used to be the first index
		// but is now outdated and references a log that does not exist.
		//
		// When FirstIndex returns an invalid index, the GetLog method
		// will be invoked with that index and return an error. Raft will
		// infinitely retry GetLog if it returns an error.
		//
		// So this code is essentially overruling what BadgerDB thinks
		// is the first index based on what we know about the compaction
		// process. This helps us avoid infinite loops in the Raft algorithm.
		index = b.maxCompactedIdx + 1
	}
	return index, nil
}

func (b *BadgerLogStore) LastIndex() (uint64, error) {
	var index uint64
	return index, b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Reverse = true
		opts.Prefix = logStorePrefix

		it := txn.NewIterator(opts)
		defer it.Close()
		it.Seek(getLogKey(math.MaxUint64))
		if !it.Valid() {
			return nil
		}

		k := it.Item().Key()
		idx := k[len(k)-8:]
		index = binary.BigEndian.Uint64(idx)
		return nil
	})
}

func (b *BadgerLogStore) GetLog(index uint64, log *raft.Log) error {
	return b.cache.Get(index, log)
}

func (b *BadgerLogStore) StoreLog(log *raft.Log) error {
	return b.db.Update(func(txn *badger.Txn) error {
		payload, err := msgpack.Marshal(log)
		if err != nil {
			return err
		}

		err = txn.Set(getLogKey(log.Index), payload)
		if err != nil {
			return err
		}
		return b.cache.Set(log.Index, payload)
	})
}

func (b *BadgerLogStore) StoreLogs(logs []*raft.Log) error {
	batch := b.db.NewWriteBatch()
	defer batch.Cancel()

	for _, log := range logs {
		payload, err := msgpack.Marshal(log)
		if err != nil {
			return err
		}

		err = batch.Set(getLogKey(log.Index), payload)
		if err != nil {
			return err
		}

		err = b.cache.Set(log.Index, payload)
		if err != nil {
			return err
		}
	}
	return batch.Flush()
}

func (b *BadgerLogStore) DeleteRange(min, max uint64) error {
	batch := b.db.NewWriteBatch()
	defer batch.Cancel()

	for i := min; i <= max; i++ {
		err := batch.Delete(getLogKey(i))
		if err != nil {
			return err
		}
		b.cache.Remove(i)
	}
	b.maxCompactedIdx = max
	return batch.Flush()
}

func getLogKey(index uint64) []byte {
	k := bytes.NewBuffer(nil)
	k.Write(logStorePrefix)
	_ = binary.Write(k, binary.BigEndian, index)
	return k.Bytes()
}
