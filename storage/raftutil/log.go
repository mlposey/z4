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
	db     *badger.DB
	prefix []byte
}

var _ raft.LogStore = (*BadgerLogStore)(nil)

func NewLogStore(db *badger.DB) (*BadgerLogStore, error) {
	return &BadgerLogStore{
		db:     db,
		prefix: []byte("raft#logstore#"),
	}, nil
}

func (b *BadgerLogStore) FirstIndex() (uint64, error) {
	var index uint64
	return index, b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()
		it.Seek(b.getKey(0))
		if !it.ValidForPrefix(b.prefix) {
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

		it := txn.NewIterator(opts)
		defer it.Close()
		it.Seek(b.getKey(math.MaxUint64))
		if !it.ValidForPrefix(b.prefix) {
			return nil
		}

		k := it.Item().Key()
		idx := k[len(k)-8:]
		index = binary.BigEndian.Uint64(idx)
		return nil
	})
}

func (b *BadgerLogStore) GetLog(index uint64, log *raft.Log) error {
	return b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(b.getKey(index))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return msgpack.Unmarshal(val, log)
		})
	})
}

func (b *BadgerLogStore) StoreLog(log *raft.Log) error {
	return b.db.Update(func(txn *badger.Txn) error {
		payload, err := msgpack.Marshal(log)
		if err != nil {
			return err
		}
		return txn.Set(b.getKey(log.Index), payload)
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
		err = batch.Set(b.getKey(log.Index), payload)
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
		err := batch.Delete(b.getKey(i))
		if err != nil {
			return err
		}
	}
	return batch.Flush()
}

func (b *BadgerLogStore) getKey(index uint64) []byte {
	k := bytes.NewBuffer(nil)
	k.Write(b.prefix)
	_ = binary.Write(k, binary.BigEndian, index)
	return k.Bytes()
}
