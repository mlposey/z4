package server

import (
	"fmt"
	"github.com/mlposey/z4/feeds"
	"github.com/mlposey/z4/feeds/q"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/server/api"
	"github.com/mlposey/z4/server/cluster"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type Config struct {
	DB          *storage.BadgerClient
	GRPCPort    int
	MetricsPort int
	PeerConfig  cluster.PeerConfig
	Opts        []grpc.ServerOption
}

type Server struct {
	fm     *feeds.Manager
	config Config
	server *grpc.Server
	peer   *cluster.Peer
	idx    *storage.IndexStore
}

func NewServer(config Config) *Server {
	return &Server{config: config}
}

func (s *Server) Start() error {
	telemetry.Logger.Info("starting server...")
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.config.GRPCPort))
	if err != nil {
		return err
	}
	telemetry.Logger.Info("listening for connections",
		zap.Int("port", s.config.GRPCPort))

	s.config.PeerConfig.Tasks = storage.NewTaskStore(s.config.DB)
	s.config.PeerConfig.Namespaces = storage.NewNamespaceStore(s.config.DB)
	s.config.PeerConfig.Writer = q.NewTaskWriter(s.config.PeerConfig.Tasks)

	s.config.PeerConfig.DB = s.config.DB
	s.peer, err = cluster.NewPeer(s.config.PeerConfig)
	if err != nil {
		return fmt.Errorf("failed to start raft server: %w", err)
	}
	s.fm = feeds.NewManager(s.config.DB, s.peer.Raft)

	tracker := cluster.NewTracker(s.peer.Raft, s.config.PeerConfig.ID)
	handle, err := cluster.NewHandle(tracker, s.config.GRPCPort)
	if err != nil {
		return fmt.Errorf("failed to obtain leader handle: %w", err)
	}
	s.peer.LoadHandle(handle)

	s.idx = storage.NewIndexStore(s.config.PeerConfig.Namespaces)
	gen := storage.NewGenerator(s.idx)

	s.server = grpc.NewServer(s.config.Opts...)
	adminServer := api.NewAdmin(s.peer.Raft, s.config.PeerConfig, handle, s.fm, gen)
	proto.RegisterAdminServer(s.server, adminServer)

	collectionServer := api.NewQueue(s.fm, s.config.PeerConfig.Tasks,
		s.peer.Raft, handle, gen)
	proto.RegisterQueueServer(s.server, collectionServer)

	go telemetry.StartPromServer(s.config.MetricsPort)
	return s.server.Serve(lis)
}

func (s *Server) Close() error {
	telemetry.Logger.Info("stopping server...")
	s.server.Stop()
	// s.server.GracefulStop() - doesnt stop streams, causing server to stay running
	err := multierr.Combine(
		s.peer.Close(),
		s.fm.Close(),
		s.config.PeerConfig.Writer.Close(),
		s.idx.Close())
	telemetry.Logger.Info("server stopped")
	return err
}
