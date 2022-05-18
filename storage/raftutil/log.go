package raftutil

import (
	"bytes"
	"encoding/binary"
	"github.com/cockroachdb/pebble"
	"github.com/hashicorp/raft"
	"github.com/vmihailenco/msgpack/v5"
	"math"
)

type BadgerLogStore struct {
	db    *pebble.DB
	cache *logCache
}

var _ raft.LogStore = (*BadgerLogStore)(nil)
var logStorePrefix = []byte("raft#logstore#")

func NewLogStore(db *pebble.DB) (*BadgerLogStore, error) {
	return &BadgerLogStore{
		db: db,
		// TODO: Find a good cache size.
		// This value seems to work well in practice, but it was
		// really just pulled from a hat.
		cache: newLogCache(db, 100_000),
	}, nil
}

func (b *BadgerLogStore) FirstIndex() (uint64, error) {
	it := b.db.NewIter(new(pebble.IterOptions))
	defer it.Close()
	found := it.SeekGE(getLogKey(0))
	if !found {
		return 0, nil
	}

	key := it.Key()
	idx := key[len(key)-8:]
	index := binary.BigEndian.Uint64(idx)

	var log marshaledLog = make([]byte, len(it.Value()))
	copy(log, it.Value())
	return index, b.cache.Set(index, log)
}

func (b *BadgerLogStore) LastIndex() (uint64, error) {
	it := b.db.NewIter(new(pebble.IterOptions))
	defer it.Close()
	found := it.SeekLT(getLogKey(math.MaxUint64))
	if !found {
		return 0, nil
	}

	key := it.Key()
	idx := key[len(key)-8:]
	index := binary.BigEndian.Uint64(idx)

	var log marshaledLog = make([]byte, len(it.Value()))
	copy(log, it.Value())
	return index, b.cache.Set(index, log)
}

func (b *BadgerLogStore) GetLog(index uint64, log *raft.Log) error {
	return b.cache.Get(index, log)
}

func (b *BadgerLogStore) StoreLog(log *raft.Log) error {
	payload, err := msgpack.Marshal(log)
	if err != nil {
		return err
	}

	err = b.db.Set(getLogKey(log.Index), payload, pebble.NoSync)
	if err != nil {
		return err
	}
	return b.cache.Set(log.Index, payload)
}

func (b *BadgerLogStore) StoreLogs(logs []*raft.Log) error {
	batch := b.db.NewBatch()
	for _, log := range logs {
		payload, err := msgpack.Marshal(log)
		if err != nil {
			_ = batch.Close()
			return err
		}

		err = batch.Set(getLogKey(log.Index), payload, pebble.NoSync)
		if err != nil {
			_ = batch.Close()
			return err
		}

		err = b.cache.Set(log.Index, payload)
		if err != nil {
			_ = batch.Close()
			return err
		}
	}
	return batch.Commit(pebble.NoSync)
}

func (b *BadgerLogStore) DeleteRange(min, max uint64) error {
	for i := min; i <= max; i++ {
		b.cache.Remove(i)
	}
	return b.db.DeleteRange(
		getLogKey(min),
		getLogKey(max+1),
		pebble.NoSync,
	)
}

func getLogKey(index uint64) []byte {
	k := bytes.NewBuffer(nil)
	k.Write(logStorePrefix)
	_ = binary.Write(k, binary.BigEndian, index)
	return k.Bytes()
}
