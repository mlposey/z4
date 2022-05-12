package storage

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/multierr"
	"time"
)

// BadgerClient manages a connection to a BadgerDB file store.
type BadgerClient struct {
	DB *badger.DB
	r  *reporter
}

func NewBadgerClient(dataDir string) (*BadgerClient, error) {
	db, err := badger.Open(badger.DefaultOptions(dataDir))
	if err != nil {
		return nil, fmt.Errorf("could not open badger database: %w", err)
	}
	client := &BadgerClient{DB: db}
	client.r = newReporter(client)
	go client.handleGC()
	go client.r.StartReporter()
	return client, nil
}

func (bc *BadgerClient) handleGC() {
	// TODO: Consider making these values configurable.

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		telemetry.Logger.Debug("running value log garbage collector")

	again:
		err := bc.DB.RunValueLogGC(0.7)
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
		bc.r.Close(),
		bc.DB.Close(),
	)
}
