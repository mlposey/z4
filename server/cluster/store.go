package cluster

import (
	"github.com/cockroachdb/pebble"
	"github.com/mlposey/z4/server/cluster/sm"
	"go.uber.org/multierr"
)

type pebbleStore struct {
	db *pebble.DB
}

func newPebbleStore(db *pebble.DB) *pebbleStore {
	return &pebbleStore{db: db}
}

func (w *pebbleStore) Write(k, v []byte) error {
	return w.db.Set(k, v, pebble.NoSync)
}

func (w *pebbleStore) DeleteAll() error {
	err := w.db.DeleteRange([]byte("task# "), []byte("task#z"), pebble.NoSync)
	if err != nil {
		return err
	}
	return w.db.DeleteRange([]byte("queue# "), []byte("queue#z"), pebble.NoSync)
}

func (w *pebbleStore) Iterator() sm.KVIterator {
	return newSnapshotScanner(w.db)
}

type snapshotScanner struct {
	db *pebble.Snapshot
}

func newSnapshotScanner(db *pebble.DB) *snapshotScanner {
	return &snapshotScanner{db: db.NewSnapshot()}
}

func (s *snapshotScanner) Close() error {
	return s.db.Close()
}

func (s *snapshotScanner) ForEach(visit func(k, v []byte) error) error {
	it := s.db.NewIter(&pebble.IterOptions{})
	it.First()

	for ; it.Valid(); it.Next() {
		err := visit(it.Key(), it.Value())
		if err != nil {
			return multierr.Combine(err, it.Close())
		}
	}
	return it.Close()
}
