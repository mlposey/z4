package raftutil

import (
	"bytes"
	"encoding/binary"
	"github.com/cockroachdb/pebble"
	"github.com/hashicorp/raft"
	"github.com/vmihailenco/msgpack/v5"
	"math"
)

type PebbleLogStore struct {
	db *pebble.DB
}

var _ raft.LogStore = (*PebbleLogStore)(nil)
var LogStorePrefix = []byte("raft#logstore#")

func NewLogStore(db *pebble.DB) (*PebbleLogStore, error) {
	return &PebbleLogStore{db: db}, nil
}

func (b *PebbleLogStore) FirstIndex() (uint64, error) {
	it := b.db.NewIter(&pebble.IterOptions{})
	defer it.Close()
	found := it.SeekGE(getLogKey(0))
	if !found || !bytes.HasPrefix(it.Key(), LogStorePrefix) {
		return 0, nil
	}

	key := it.Key()
	idx := key[len(key)-8:]
	return binary.BigEndian.Uint64(idx), nil
}

func (b *PebbleLogStore) LastIndex() (uint64, error) {
	it := b.db.NewIter(&pebble.IterOptions{})
	defer it.Close()
	found := it.SeekLT(getLogKey(math.MaxInt64))
	if !found || !bytes.HasPrefix(it.Key(), LogStorePrefix) {
		return 0, nil
	}

	key := it.Key()
	idx := key[len(key)-8:]
	return binary.BigEndian.Uint64(idx), nil
}

func (b *PebbleLogStore) GetLog(index uint64, log *raft.Log) error {
	val, closer, err := b.db.Get(getLogKey(index))
	if err != nil {
		return err
	}

	err = msgpack.Unmarshal(val, log)
	if err != nil {
		return err
	}
	return closer.Close()
}

func (b *PebbleLogStore) StoreLog(log *raft.Log) error {
	payload, err := msgpack.Marshal(log)
	if err != nil {
		return err
	}

	return b.db.Set(getLogKey(log.Index), payload, pebble.NoSync)
}

func (b *PebbleLogStore) StoreLogs(logs []*raft.Log) error {
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
	}
	return batch.Commit(pebble.NoSync)
}

func (b *PebbleLogStore) DeleteRange(min, max uint64) error {
	return b.db.DeleteRange(
		getLogKey(min),
		getLogKey(max+1),
		pebble.NoSync,
	)
}

func getLogKey(index uint64) []byte {
	k := bytes.NewBuffer(nil)
	k.Write(LogStorePrefix)
	_ = binary.Write(k, binary.BigEndian, index)
	return k.Bytes()
}
