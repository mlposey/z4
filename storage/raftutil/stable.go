package raftutil

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
)

type BadgerStableStore struct {
	db     *badger.DB
	prefix []byte
}

var _ raft.StableStore = (*BadgerStableStore)(nil)

func NewStableStore(db *badger.DB) *BadgerStableStore {
	return &BadgerStableStore{
		db:     db,
		prefix: []byte("raft#stablestore#"),
	}
}

func (b *BadgerStableStore) Set(key []byte, val []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(b.getKey(key), val)
	})
}

func (b *BadgerStableStore) Get(key []byte) ([]byte, error) {
	var out []byte
	return out, b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(b.getKey(key))
		if err != nil {
			return err
		}

		out, err = item.ValueCopy(nil)
		return err
	})
}

func (b *BadgerStableStore) SetUint64(key []byte, val uint64) error {
	encoded := make([]byte, 8)
	binary.BigEndian.PutUint64(encoded, val)
	return b.Set(b.getKey(key), encoded)
}

func (b *BadgerStableStore) GetUint64(key []byte) (uint64, error) {
	val, err := b.Get(b.getKey(key))
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
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
