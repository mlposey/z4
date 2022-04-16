package cluster

import (
	"errors"
	"fmt"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"net"
	"os"
	"path/filepath"
)

// PeerConfig defines how a node will take part in the raft cluster.
type PeerConfig struct {
	ID               string
	Port             int
	AdvertiseAddr    string
	DataDir          string
	LogBatchSize     int
	BootstrapCluster bool
	DB               *storage.BadgerClient
	Tasks            *storage.TaskStore
}

// Peer controls a node's membership in the raft cluster.
type Peer struct {
	Raft        *raft.Raft
	addr        string
	config      PeerConfig
	logStore    *boltdb.BoltStore
	stableStore *boltdb.BoltStore
	snapshots   *raft.FileSnapshotStore
	transport   *raft.NetworkTransport
}

func NewPeer(config PeerConfig) (*Peer, error) {
	peer := &Peer{config: config}

	err := peer.initStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to load raft storage: %w", err)
	}

	err = peer.joinNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to join raft network: %w", err)
	}

	peer.tryBootstrap()
	return peer, nil
}

func (p *Peer) initStorage() error {
	_, err := os.Stat(p.config.DataDir)
	if errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(p.config.DataDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create peer storage folder: %w", err)
		}
	}

	logStorePath := filepath.Join(p.config.DataDir, "logs.dat")
	p.logStore, err = boltdb.NewBoltStore(logStorePath)
	if err != nil {
		return fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, logStorePath, err)
	}

	stableStorePath := filepath.Join(p.config.DataDir, "stable.dat")
	p.stableStore, err = boltdb.NewBoltStore(stableStorePath)
	if err != nil {
		return fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, stableStorePath, err)
	}

	p.snapshots, err = raft.NewFileSnapshotStore(p.config.DataDir, 3, os.Stderr)
	if err != nil {
		return fmt.Errorf(`raft.NewFileSnapshotStore(%q, ...): %v`, p.config.DataDir, err)
	}
	return nil
}

func (p *Peer) joinNetwork() error {
	c := raft.DefaultConfig()
	c.LocalID = raft.ServerID(p.config.ID)

	c.BatchApplyCh = true
	c.MaxAppendEntries = p.config.LogBatchSize

	p.addr = fmt.Sprintf("0.0.0.0:%d", p.config.Port)
	advertise, err := net.ResolveTCPAddr("tcp", p.config.AdvertiseAddr)
	if err != nil {
		return fmt.Errorf("failed to parse avertise address: %s: %w", p.config.AdvertiseAddr, err)
	}

	p.transport, err = raft.NewTCPTransport(p.addr, advertise, 0, 0, nil)
	if err != nil {
		return fmt.Errorf("could not create transport for raft peer: %w", err)
	}

	fsm := newFSM(p.config.DB.DB, p.config.Tasks)
	p.Raft, err = raft.NewRaft(
		c,
		fsm,
		p.logStore,
		p.stableStore,
		p.snapshots,
		p.transport)
	if err != nil {
		return fmt.Errorf("raft.NewRaft: %v", err)
	}
	return nil
}

func (p *Peer) tryBootstrap() {
	if !p.config.BootstrapCluster {
		return
	}
	telemetry.Logger.Info("bootstrapping cluster")

	cfg := raft.Configuration{
		Servers: []raft.Server{
			{
				Suffrage: raft.Voter,
				ID:       raft.ServerID(p.config.ID),
				Address:  raft.ServerAddress(p.addr),
			},
		},
	}
	f := p.Raft.BootstrapCluster(cfg)
	if err := f.Error(); err != nil {
		telemetry.Logger.Error("failed to bootstrap cluster",
			zap.Error(err))
	}
}

// Close stops the raft server and flushes writes to disk.
//
// This method must be called before the application terminates.
func (p *Peer) Close() error {
	return multierr.Combine(
		p.Raft.Shutdown().Error(),
		p.transport.Close(),
		p.logStore.Close(),
		p.stableStore.Close())
}
