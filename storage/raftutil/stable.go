package raftutil

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/cockroachdb/pebble"
	"github.com/hashicorp/raft"
)

type BadgerStableStore struct {
	db     *pebble.DB
	prefix []byte
}

var _ raft.StableStore = (*BadgerStableStore)(nil)

func NewStableStore(db *pebble.DB) *BadgerStableStore {
	return &BadgerStableStore{
		db:     db,
		prefix: []byte("raft#stablestore#"),
	}
}

func (b *BadgerStableStore) Set(key []byte, val []byte) error {
	return b.db.Set(b.getKey(key), val, pebble.NoSync)
}

func (b *BadgerStableStore) Get(key []byte) ([]byte, error) {
	val, closer, err := b.db.Get(b.getKey(key))
	if err != nil {
		return nil, err
	}
	res := make([]byte, len(val))
	copy(res, val)
	return res, closer.Close()
}

func (b *BadgerStableStore) SetUint64(key []byte, val uint64) error {
	encoded := make([]byte, 8)
	binary.BigEndian.PutUint64(encoded, val)
	return b.Set(b.getKey(key), encoded)
}

func (b *BadgerStableStore) GetUint64(key []byte) (uint64, error) {
	val, err := b.Get(b.getKey(key))
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return binary.BigEndian.Uint64(val), nil
}

func (b *BadgerStableStore) getKey(key []byte) []byte {
	k := bytes.NewBuffer(nil)
	k.Write(b.prefix)
	k.Write(key)
	return k.Bytes()
}
