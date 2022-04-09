package server

import (
	"errors"
	"fmt"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/multierr"
	"os"
	"path/filepath"
)

// PeerConfig defines how a node will take part in the raft cluster.
type PeerConfig struct {
	ID               string
	Port             int
	DataDir          string
	LogBatchSize     int
	BootstrapCluster bool
	DB               *storage.BadgerClient
	Tasks            *storage.TaskStore
}

// raftPeer controls a node's membership in the raft cluster.
type raftPeer struct {
	Raft        *raft.Raft
	addr        string
	config      PeerConfig
	logStore    *boltdb.BoltStore
	stableStore *boltdb.BoltStore
	snapshots   *raft.FileSnapshotStore
	transport   *raft.NetworkTransport
}

func newPeer(config PeerConfig) (*raftPeer, error) {
	peer := &raftPeer{config: config}

	err := peer.initStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to load raft storage: %w", err)
	}

	err = peer.joinNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to join raft network: %w", err)
	}

	err = peer.tryBootstrap()
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap cluster: %w", err)
	}
	return peer, nil
}

func (rp *raftPeer) initStorage() error {
	_, err := os.Stat(rp.config.DataDir)
	if errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(rp.config.DataDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create raft raftPeer folder: %w", err)
		}
	}

	logStorePath := filepath.Join(rp.config.DataDir, "logs.dat")
	rp.logStore, err = boltdb.NewBoltStore(logStorePath)
	if err != nil {
		return fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, logStorePath, err)
	}

	stableStorePath := filepath.Join(rp.config.DataDir, "stable.dat")
	rp.stableStore, err = boltdb.NewBoltStore(stableStorePath)
	if err != nil {
		return fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, stableStorePath, err)
	}

	rp.snapshots, err = raft.NewFileSnapshotStore(rp.config.DataDir, 3, os.Stderr)
	if err != nil {
		return fmt.Errorf(`raft.NewFileSnapshotStore(%q, ...): %v`, rp.config.DataDir, err)
	}
	return nil
}

func (rp *raftPeer) joinNetwork() error {
	c := raft.DefaultConfig()
	c.LocalID = raft.ServerID(rp.config.ID)

	c.BatchApplyCh = true
	c.MaxAppendEntries = rp.config.LogBatchSize

	rp.addr = fmt.Sprintf("127.0.0.1:%d", rp.config.Port)
	var err error
	rp.transport, err = raft.NewTCPTransport(rp.addr, nil, 0, 0, nil)
	if err != nil {
		return fmt.Errorf("could not create transport for raft raftPeer: %w", err)
	}

	fsm := newFSM(rp.config.DB.DB, rp.config.Tasks)
	rp.Raft, err = raft.NewRaft(
		c,
		fsm,
		rp.logStore,
		rp.stableStore,
		rp.snapshots,
		rp.transport)
	if err != nil {
		return fmt.Errorf("raft.NewRaft: %v", err)
	}
	return nil
}

func (rp *raftPeer) tryBootstrap() error {
	if !rp.config.BootstrapCluster {
		return nil
	}
	telemetry.Logger.Info("bootstrapping cluster")

	cfg := raft.Configuration{
		Servers: []raft.Server{
			{
				Suffrage: raft.Voter,
				ID:       raft.ServerID(rp.config.ID),
				Address:  raft.ServerAddress(rp.addr),
			},
		},
	}
	f := rp.Raft.BootstrapCluster(cfg)
	if err := f.Error(); err != nil {
		return fmt.Errorf("raft.Raft.BootstrapCluster: %v", err)
	}
	return nil
}

// Close stops the raft server and flushes writes to disk.
//
// This method must be called before the application terminates.
func (rp *raftPeer) Close() error {
	return multierr.Combine(
		rp.transport.Close(),
		rp.logStore.Close(),
		rp.stableStore.Close())
}
