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
	peb, err := pebble.Open(dataDir, &pebble.Options{
		MemTableSize: 100 << 20,
	})
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
