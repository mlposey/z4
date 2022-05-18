package storage

import (
	"fmt"
	"github.com/cockroachdb/pebble"
	"github.com/dgraph-io/badger/v3"
	"go.uber.org/multierr"
)

// BadgerClient manages a connection to a BadgerDB file store.
type BadgerClient struct {
	DB  *badger.DB
	DB2 *pebble.DB
}

func NewBadgerClient(dataDir string) (*BadgerClient, error) {
	opts := badger.DefaultOptions(dataDir)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("could not open badger database: %w", err)
	}

	peb, err := pebble.Open(dataDir+"/pebble", &pebble.Options{
		MemTableSize: 100 << 20,
		//LBaseMaxBytes: 100 << 20,
	})
	if err != nil {
		return nil, fmt.Errorf("could not open pebble database: %w", err)
	}

	return &BadgerClient{
		DB:  db,
		DB2: peb,
	}, nil
}

// Close flushes writes to disk.
//
// It must be called when the client is no longer needed or else
// pending writes may be canceled when the application terminates.
func (bc *BadgerClient) Close() error {
	return multierr.Combine(
		bc.DB2.Close(),
		bc.DB.Close(),
	)
}
