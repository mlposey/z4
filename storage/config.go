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

// SyncedNamespace is a Namespace that syncs to disk.
type SyncedNamespace struct {
	// N is the namespace that will be synced to disk.
	// It is safe to make changes directly to this field.
	N *proto.Namespace

	namespaces *NamespaceStore
	lastSaved  *proto.Namespace
	namespace  string
	closeReq   chan interface{}
	closeRes   chan interface{}
}

func NewSyncedNamespace(namespaces *NamespaceStore, namespace string) *SyncedNamespace {
	return &SyncedNamespace{
		namespaces: namespaces,
		namespace:  namespace,
		closeReq:   make(chan interface{}),
		closeRes:   make(chan interface{}),
	}
}

func (sn *SyncedNamespace) StartSync() error {
	if err := sn.load(); err != nil {
		return err
	}

	go sn.startSync()
	return nil
}

// Close flushes config changes to disk and stops the sync thread.
// The object should not be used after a call to Close.
func (sn *SyncedNamespace) Close() error {
	sn.closeReq <- nil
	<-sn.closeRes
	return nil
}

// load pulls the namespace from the database.
func (sn *SyncedNamespace) load() error {
	namespace, err := sn.namespaces.Get(sn.namespace)
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("failed to load namespace config from database: %w", err)
		}

		namespace = &proto.Namespace{
			Id:                sn.namespace,
			LastDeliveredTask: NewTaskID(ksuid.Nil.Time()),
		}
		err = sn.namespaces.Save(namespace)
		if err != nil {
			return fmt.Errorf("failed to save namespace to database: %w", err)
		}
	}
	sn.N = namespace
	return nil
}

// startSync starts a loop that flushes changes to disk on an interval.
func (sn *SyncedNamespace) startSync() {
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-sn.closeReq:
			err := sn.trySave()
			if err != nil {
				telemetry.Logger.Error("failed to save config to database",
					zap.Error(err))
			}
			sn.closeRes <- nil
			return

		// TODO: Consider syncing after X number of changes if before tick.

		case <-ticker.C:
			err := sn.trySave()
			if err != nil {
				telemetry.Logger.Error("failed to save config to database",
					zap.Error(err))
			}
		}
	}
}

// trySave saves the config if it has changed since the last write.
func (sn *SyncedNamespace) trySave() error {
	if pb.Equal(sn.N, sn.lastSaved) {
		return nil
	}

	snapshot := pb.Clone(sn.N).(*proto.Namespace)
	err := sn.namespaces.Save(snapshot)
	if err == nil {
		sn.lastSaved = snapshot
	}
	return err
}
