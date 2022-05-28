package storage

import (
	"errors"
	"fmt"
	"github.com/cockroachdb/pebble"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/iden"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	pb "google.golang.org/protobuf/proto"
	"time"
)

// SyncedSettings are queue settings that are synchronized to disk.
type SyncedSettings struct {
	// S is the queue settings that will be synced to disk.
	// It is safe to make changes directly to this field.
	S *proto.QueueConfig

	store     *SettingStore
	lastSaved *proto.QueueConfig
	queue     string
	closeReq  chan interface{}
	closeRes  chan interface{}
	raft      *raft.Raft
}

func NewSyncedSettings(
	store *SettingStore,
	queue string,
	raft *raft.Raft,
) *SyncedSettings {
	return &SyncedSettings{
		store:    store,
		queue:    queue,
		closeReq: make(chan interface{}),
		closeRes: make(chan interface{}),
		raft:     raft,
	}
}

func (s *SyncedSettings) StartSync() error {
	if err := s.load(); err != nil {
		return err
	}

	go s.startSync()
	return nil
}

// Close flushes config changes to disk and stops the sync thread.
// The object should not be used after a call to Close.
func (s *SyncedSettings) Close() error {
	s.closeReq <- nil
	<-s.closeRes
	return nil
}

// load pulls the settings from the database.
func (s *SyncedSettings) load() error {
	settings, err := s.store.Get(s.queue)
	if err == nil {
		s.S = settings
		return nil
	}

	if !errors.Is(err, pebble.ErrNotFound) {
		return fmt.Errorf("failed to load queue settings from database: %w", err)
	}

	// Default settings for new store go here.
	s.S = &proto.QueueConfig{
		Id:                         s.queue,
		LastDeliveredQueuedTask:    iden.Max.String(),
		LastDeliveredScheduledTask: iden.Min.String(),
		AckDeadlineSeconds:         300, // 5 minutes
	}

	err = s.trySave()
	if err != nil {
		return fmt.Errorf("failed to save queue settings to database: %w", err)
	}
	return nil
}

// startSync starts a loop that flushes changes to disk on an interval.
func (s *SyncedSettings) startSync() {
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-s.closeReq:
			err := s.trySave()
			if err != nil {
				telemetry.Logger.Error("failed to save queue settings to database",
					zap.Error(err))
			}
			s.closeRes <- nil
			return

		// TODO: Consider syncing after X number of changes if before tick.

		case <-ticker.C:
			err := s.trySave()
			if err != nil {
				telemetry.Logger.Error("failed to save queue settings to database",
					zap.Error(err))
			}
		}
	}
}

// trySave saves the config if it has changed since the last write.
func (s *SyncedSettings) trySave() error {
	if pb.Equal(s.S, s.lastSaved) {
		return nil
	}

	snapshot := pb.Clone(s.S).(*proto.QueueConfig)
	cmd, _ := pb.Marshal(&proto.Command{
		Cmd: &proto.Command_Queue{
			Queue: snapshot,
		},
	})

	err := s.raft.Apply(cmd, 0).Error()
	if err == nil {
		s.lastSaved = snapshot
	}
	return err
}
