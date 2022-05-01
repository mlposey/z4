package storage

import (
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/telemetry"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	pb "google.golang.org/protobuf/proto"
	"time"
)

// SyncedConfig is a FeedConfig that syncs to disk.
type SyncedConfig struct {
	// C is the config that will be synced to disk.
	// It is safe to make changes directly to this field.
	C *proto.Namespace

	configs   *ConfigStore
	lastSaved *proto.Namespace
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

func (sc *SyncedConfig) StartSync() error {
	if err := sc.loadConfig(); err != nil {
		return err
	}

	go sc.startConfigSync()
	return nil
}

// Close flushes config changes to disk and stops the sync thread.
// The object should not be used after a call to Close.
func (sc *SyncedConfig) Close() error {
	sc.closeReq <- nil
	<-sc.closeRes
	return nil
}

func (sc *SyncedConfig) loadConfig() error {
	config, err := sc.configs.Get(sc.namespace)
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("failed to load namespace config from database: %w", err)
		}

		config = &proto.Namespace{
			LastDeliveredTask: NewTaskID(ksuid.Nil.Time()),
		}
		err = sc.configs.Save(sc.namespace, config)
		if err != nil {
			return fmt.Errorf("failed to save config to database: %w", err)
		}
	}
	sc.C = config
	return nil
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
	if pb.Equal(sc.C, sc.lastSaved) {
		return nil
	}

	snapshot := pb.Clone(sc.C).(*proto.Namespace)
	err := sc.configs.Save(sc.namespace, snapshot)
	if err == nil {
		sc.lastSaved = snapshot
	}
	return err
}
