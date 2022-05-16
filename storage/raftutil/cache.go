package raftutil

import (
	"github.com/bluele/gcache"
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"github.com/vmihailenco/msgpack/v5"
)

type logCache struct {
	db   *badger.DB
	logs gcache.Cache
}

type marshaledLog []byte

func newLogCache(db *badger.DB, size int) *logCache {
	p := &logCache{
		db: db,
	}
	p.logs = gcache.New(size).
		LRU().
		LoaderFunc(func(i interface{}) (interface{}, error) {
			return p.load(i.(uint64))
		}).
		Build()
	return p
}

func (p *logCache) Set(index uint64, log marshaledLog) error {
	return p.logs.Set(index, log)
}

func (p *logCache) Get(index uint64, log *raft.Log) error {
	um, err := p.logs.Get(index)
	if err != nil {
		return err
	}
	return msgpack.Unmarshal(um.(marshaledLog), log)
}

func (p *logCache) Remove(index uint64) {
	p.logs.Remove(index)
}

func (p *logCache) load(index uint64) (marshaledLog, error) {
	// TODO: Record cache misses using a metric.
	var log marshaledLog
	return log, p.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(getLogKey(index))
		if err != nil {
			return err
		}

		log, err = item.ValueCopy(nil)
		return err
	})
}
