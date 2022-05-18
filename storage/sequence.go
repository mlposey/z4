package storage

import (
	"encoding/binary"
	"errors"
	"github.com/bluele/gcache"
	"github.com/cockroachdb/pebble"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"sync"
)

type Sequence struct {
	db        *pebble.DB
	leaseSize uint64
	nextLease uint64
	i         uint64
	key       []byte
	mu        sync.Mutex
	buf       []byte
}

func NewSequence(db *pebble.DB, key []byte, lease uint64) (*Sequence, error) {
	s := &Sequence{
		db:        db,
		leaseSize: lease,
		i:         0,
		key:       key,
		buf:       make([]byte, 8),
	}
	return s, s.init()
}

func (s *Sequence) init() error {
	val, closer, err := s.db.Get(s.key)
	if err != nil {
		if !errors.Is(err, pebble.ErrNotFound) {
			return err
		}
		s.i = 0
	} else {
		defer closer.Close()
		s.i = binary.BigEndian.Uint64(val)
	}
	return s.reserve()
}

func (s *Sequence) reserve() error {
	s.nextLease = s.i + s.leaseSize - 1
	binary.BigEndian.PutUint64(s.buf, s.nextLease)
	return s.db.Set(s.key, s.buf, pebble.NoSync)
}

func (s *Sequence) Next() (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.i == s.nextLease {
		err := s.reserve()
		if err != nil {
			return 0, err
		}
	}
	s.i++
	return s.i, nil
}

func (s *Sequence) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	binary.BigEndian.PutUint64(s.buf, s.i)
	return s.db.Set(s.key, s.buf, pebble.NoSync)
}

type IndexStore struct {
	seqCache gcache.Cache
}

func NewIndexStore(namespaces *NamespaceStore) *IndexStore {
	release := func(namespace string, seq *Sequence) {
		telemetry.Logger.Info("closing seq")
		err := seq.Close()
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
				return namespaces.Sequence(i.(string))
			}).
			EvictedFunc(func(key interface{}, value interface{}) {
				release(key.(string), value.(*Sequence))
			}).
			PurgeVisitorFunc(func(key interface{}, value interface{}) {
				release(key.(string), value.(*Sequence))
			}).
			Build(),
	}
}

func (c *IndexStore) Next(namespace string) (uint64, error) {
	seq, err := c.seqCache.Get(namespace)
	if err != nil {
		return 0, err
	}
	return seq.(*Sequence).Next()
}

func (c *IndexStore) Close() error {
	c.seqCache.Purge()
	return nil
}
