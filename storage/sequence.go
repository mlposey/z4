package storage

import (
	"github.com/bluele/gcache"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
)

type IndexStore struct {
	seqCache gcache.Cache
}

func NewIndexStore(namespaces *NamespaceStore) *IndexStore {
	release := func(namespace string, seq *badger.Sequence) {
		telemetry.Logger.Info("closing seq")
		err := seq.Release()
		if err != nil {
			telemetry.Logger.Error("failed to close sequence",
				zap.Error(err),
				zap.String("namespace", namespace))
		}
	}

	return &IndexStore{
		seqCache: gcache.New(100).
			ARC().
			LoaderFunc(func(i interface{}) (interface{}, error) {
				// TODO: Initialize sequence if needed
				return namespaces.Sequence(i.(string))
			}).
			EvictedFunc(func(key interface{}, value interface{}) {
				release(key.(string), value.(*badger.Sequence))
			}).
			PurgeVisitorFunc(func(key interface{}, value interface{}) {
				release(key.(string), value.(*badger.Sequence))
			}).
			Build(),
	}
}

func (c *IndexStore) Next(namespace string) (uint64, error) {
	seq, err := c.seqCache.Get(namespace)
	if err != nil {
		return 0, err
	}
	return seq.(*badger.Sequence).Next()
}

func (c *IndexStore) Close() error {
	c.seqCache.Purge()
	return nil
}
