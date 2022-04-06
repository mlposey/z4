package server

import (
	"errors"
	"fmt"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	"github.com/mlposey/z4/feeds"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"os"
	"path/filepath"
)

type Config struct {
	DB            *storage.BadgerClient
	ServicePort   int
	PeerPort      int
	PeerID        string
	Opts          []grpc.ServerOption
	RaftDataDir   string
	BootstrapRaft bool
}

type Server struct {
	tasks       *storage.TaskStore
	fm          *feeds.Manager
	fsm         *stateMachine
	config      Config
	server      *grpc.Server
	raft        *raft.Raft
	peerNetwork *raft.NetworkTransport
}

func NewServer(config Config) *Server {
	return &Server{config: config}
}

func (s *Server) Start() error {
	telemetry.Logger.Info("starting server...")
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.config.ServicePort))
	if err != nil {
		return err
	}
	telemetry.Logger.Info("listening for connections",
		zap.Int("port", s.config.ServicePort))

	s.fm = feeds.NewManager(s.config.DB)
	s.tasks = storage.NewTaskStore(s.config.DB)
	s.raft, err = s.newRaft()
	if err != nil {
		return fmt.Errorf("failed to start raft server: %w", err)
	}

	s.server = grpc.NewServer(s.config.Opts...)
	proto.RegisterAdminServer(s.server, newAdmin(s.raft, s.config.PeerID))

	proto.RegisterCollectionServer(s.server, newCollection(s.fm, s.raft))
	return s.server.Serve(lis)
}

func (s *Server) newRaft() (*raft.Raft, error) {
	// TODO: Refactor this method into a new type for raft/peer logic.

	c := raft.DefaultConfig()
	c.LocalID = raft.ServerID(s.config.PeerID)

	// TODO: Verify impact of these settings on write durability.
	// These settings greatly improve performance, but may introduce issues.
	c.BatchApplyCh = true
	c.MaxAppendEntries = 1000

	// TODO: Create z4peer folder if it does not exist.
	// Bolt is creating the logs.dat and stable.dat files but not the parent folder.
	if _, err := os.Stat(s.config.RaftDataDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(s.config.RaftDataDir, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("failed to create raft peer folder: %w", err)
		}
	}

	ldb, err := boltdb.NewBoltStore(filepath.Join(s.config.RaftDataDir, "logs.dat"))
	if err != nil {
		return nil, fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(s.config.RaftDataDir, "logs.dat"), err)
	}

	sdb, err := boltdb.NewBoltStore(filepath.Join(s.config.RaftDataDir, "stable.dat"))
	if err != nil {
		return nil, fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(s.config.RaftDataDir, "stable.dat"), err)
	}

	fss, err := raft.NewFileSnapshotStore(s.config.RaftDataDir, 3, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf(`raft.NewFileSnapshotStore(%q, ...): %v`, s.config.RaftDataDir, err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", s.config.PeerPort)
	tm, err := raft.NewTCPTransport(addr, nil, 0, 0, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create transport for raft peer: %w", err)
	}
	s.peerNetwork = tm

	s.fsm = newFSM(s.config.DB.DB, s.tasks)
	r, err := raft.NewRaft(c, s.fsm, ldb, sdb, fss, tm)
	if err != nil {
		return nil, fmt.Errorf("raft.NewRaft: %v", err)
	}

	if s.config.BootstrapRaft {
		telemetry.Logger.Info("bootstrapping cluster")
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(s.config.PeerID),
					Address:  raft.ServerAddress(addr),
				},
			},
		}
		f := r.BootstrapCluster(cfg)
		if err := f.Error(); err != nil {
			return nil, fmt.Errorf("raft.Raft.BootstrapCluster: %v", err)
		}
	}
	return r, nil
}

func (s *Server) Close() error {
	// TODO: Consider supporting a context with timeout.
	// TODO: Close raft Bolt databases.

	telemetry.Logger.Info("stopping server...")
	s.server.GracefulStop()
	err := s.peerNetwork.Close()
	err = multierr.Append(err, s.fm.Close())
	telemetry.Logger.Info("server stopped")
	return err
}
