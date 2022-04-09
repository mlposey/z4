package server

import (
	"fmt"
	"github.com/mlposey/z4/feeds"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type Config struct {
	DB         *storage.BadgerClient
	GRPCPort   int
	PeerConfig PeerConfig
	Opts       []grpc.ServerOption
}

type Server struct {
	fm     *feeds.Manager
	config Config
	server *grpc.Server
	peer   *raftPeer
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
	s.config.PeerConfig.DB = s.config.DB
	s.peer, err = newPeer(s.config.PeerConfig)
	if err != nil {
		return fmt.Errorf("failed to start raft server: %w", err)
	}

	s.server = grpc.NewServer(s.config.Opts...)
	proto.RegisterAdminServer(s.server, newAdmin(s.peer.Raft, s.config.PeerConfig.ID))

	s.fm = feeds.NewManager(s.config.DB)
	proto.RegisterCollectionServer(s.server, newCollection(s.fm, s.config.PeerConfig.Tasks, s.peer.Raft))
	return s.server.Serve(lis)
}

func (s *Server) Close() error {
	telemetry.Logger.Info("stopping server...")
	s.server.Stop()
	// s.server.GracefulStop() - doesnt stop streams, causing server to stay running
	err := multierr.Combine(
		s.peer.Close(),
		s.fm.Close())
	telemetry.Logger.Info("server stopped")
	return err
}
