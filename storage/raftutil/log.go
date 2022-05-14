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
	db    *badger.DB
	cache *logCache
}

var _ raft.LogStore = (*BadgerLogStore)(nil)
var logStorePrefix = []byte("raft#logstore#")

func NewLogStore(db *badger.DB, prefetch int) (*BadgerLogStore, error) {
	return &BadgerLogStore{
		db:    db,
		cache: newLogCache(db, prefetch),
	}, nil
}

func (b *BadgerLogStore) FirstIndex() (uint64, error) {
	var index uint64
	return index, b.db.View(func(txn *badger.Txn) error {
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
		return txn.Set(getLogKey(log.Index), payload)
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
	}
	return batch.Flush()
}

func getLogKey(index uint64) []byte {
	k := bytes.NewBuffer(nil)
	k.Write(logStorePrefix)
	_ = binary.Write(k, binary.BigEndian, index)
	return k.Bytes()
}