package storage

import (
	"fmt"
	"github.com/cockroachdb/pebble"
	"go.uber.org/multierr"
)

// PebbleClient manages a connection to a pebble file store.
type PebbleClient struct {
	DB *pebble.DB
}

func NewPebbleClient(dataDir string) (*PebbleClient, error) {
	opt := (&pebble.Options{
		L0CompactionThreshold:       2,
		L0StopWritesThreshold:       1000,
		LBaseMaxBytes:               64 << 20,
		MaxConcurrentCompactions:    3,
		MemTableSize:                64 << 20,
		MemTableStopWritesThreshold: 4,
	}).EnsureDefaults()
	for i := range opt.Levels {
		opt.Levels[i].Compression = pebble.ZstdCompression
	}

	peb, err := pebble.Open(dataDir, opt)
	if err != nil {
		return nil, fmt.Errorf("could not open pebble database: %w", err)
	}

	return &PebbleClient{
		DB: peb,
	}, nil
}

// Close flushes writes to disk.
//
// It must be called when the client is no longer needed or else
// pending writes may be canceled when the application terminates.
func (bc *PebbleClient) Close() error {
	return multierr.Combine(
		bc.DB.Close(),
	)
}
