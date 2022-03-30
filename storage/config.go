package storage

import (
	"errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"time"
)

type QueueConfig struct {
	LastRun time.Time
}

type SyncedConfig struct {
	configs   *ConfigStore
	C         *QueueConfig
	namespace string
	closed    chan bool
}

func NewSyncedConfig(configs *ConfigStore, namespace string) *SyncedConfig {
	return &SyncedConfig{
		configs:   configs,
		namespace: namespace,
		closed:    make(chan bool),
	}
}

func (sc *SyncedConfig) StartSync() {
	sc.loadConfig()
	go sc.startConfigSync()
}

func (sc *SyncedConfig) Close() error {
	sc.closed <- true
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

		config = QueueConfig{LastRun: time.Now().Add(-time.Minute)}
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
		case <-sc.closed:
			return

		// TODO: Consider syncing after X number of changes if before tick.

		case <-ticker.C:
			err := sc.configs.Save(sc.namespace, *sc.C)
			if err != nil {
				telemetry.Logger.Error("failed to save config to database",
					zap.Error(err))
			}
		}
	}
}
