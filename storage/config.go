package storage

import (
	"errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"time"
)

// FeedConfig defines how a feed should work.
type FeedConfig struct {
	// Update Equals and Copy methods if modifying fields in this type.

	LastDeliveredTask string // The ID of the last delivered task
}

// Equals compares two feeds for equality.
func (c *FeedConfig) Equals(other *FeedConfig) bool {
	if other == nil {
		return false
	}
	return c.LastDeliveredTask == other.LastDeliveredTask
}

// Copy makes a deep copy of the config.
func (c *FeedConfig) Copy() *FeedConfig {
	return &FeedConfig{
		LastDeliveredTask: c.LastDeliveredTask,
	}
}

// SyncedConfig is a FeedConfig that syncs to disk.
type SyncedConfig struct {
	// C is the config that will be synced to disk.
	// It is safe to make changes directly to this field.
	C *FeedConfig

	configs   *ConfigStore
	lastSaved *FeedConfig
	namespace string
	closeReq  chan interface{}
	closeRes  chan interface{}
}

func NewSyncedConfig(configs *ConfigStore, namespace string) *SyncedConfig {
	return &SyncedConfig{
		configs:   configs,
		namespace: namespace,
		closeReq:  make(chan interface{}),
		closeRes:  make(chan interface{}),
	}
}

func (sc *SyncedConfig) StartSync() {
	sc.loadConfig()
	go sc.startConfigSync()
}

// Close flushes config changes to disk and stops the sync thread.
// The object should not be used after a call to Close.
func (sc *SyncedConfig) Close() error {
	sc.closeReq <- nil
	<-sc.closeRes
	return nil
}

func (sc *SyncedConfig) loadConfig() {
	// TODO: Return error instead of fatal logging.
	config, err := sc.configs.Get(sc.namespace)
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			telemetry.Logger.Fatal("failed to load namespace config from database",
				zap.Error(err))
		}

		config = FeedConfig{
			LastDeliveredTask: NewTaskID(time.Now().Add(-time.Minute)),
		}
		err = sc.configs.Save(sc.namespace, config)
		if err != nil {
			telemetry.Logger.Fatal("failed to save config to database",
				zap.Error(err))
		}
	}
	sc.C = &config
}

func (sc *SyncedConfig) startConfigSync() {
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-sc.closeReq:
			err := sc.trySave()
			if err != nil {
				telemetry.Logger.Error("failed to save config to database",
					zap.Error(err))
			}
			sc.closeRes <- nil
			return

		// TODO: Consider syncing after X number of changes if before tick.

		case <-ticker.C:
			err := sc.trySave()
			if err != nil {
				telemetry.Logger.Error("failed to save config to database",
					zap.Error(err))
			}
		}
	}
}

// trySave saves the config if it has changed since the last write.
func (sc *SyncedConfig) trySave() error {
	if sc.C.Equals(sc.lastSaved) {
		return nil
	}

	snapshot := sc.C.Copy()
	err := sc.configs.Save(sc.namespace, *snapshot)
	if err == nil {
		sc.lastSaved = snapshot
	}
	return err
}
