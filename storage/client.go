package storage

import (
	"fmt"
	"github.com/cockroachdb/pebble"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/multierr"
	"time"
)

// BadgerClient manages a connection to a BadgerDB file store.
type BadgerClient struct {
	DB  *badger.DB
	DB2 *pebble.DB
	r   *reporter
}

func NewBadgerClient(dataDir string) (*BadgerClient, error) {
	opts := badger.DefaultOptions(dataDir)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("could not open badger database: %w", err)
	}

	peb, err := pebble.Open(dataDir+"/pebble", &pebble.Options{
		MemTableSize:  100 << 20,
		LBaseMaxBytes: 100 << 20,
	})
	if err != nil {
		return nil, fmt.Errorf("could not open pebble database: %w", err)
	}

	client := &BadgerClient{
		DB:  db,
		DB2: peb,
	}
	client.r = newReporter(client)
	go client.handleGC()
	go client.r.StartReporter()
	return client, nil
}

func (bc *BadgerClient) handleGC() {
	// TODO: Consider making these values configurable.

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		telemetry.Logger.Debug("running value log garbage collector")

	again:
		err := bc.DB.RunValueLogGC(0.2)
		if err == nil {
			goto again
		}
		telemetry.LastDBGC.SetToCurrentTime()
	}
}

// Close flushes writes to disk.
//
// It must be called when the client is no longer needed or else
// pending writes may be canceled when the application terminates.
func (bc *BadgerClient) Close() error {
	return multierr.Combine(
		bc.DB2.Close(),
		bc.r.Close(),
		bc.DB.Close(),
	)
}
