package server

import (
	"fmt"
	"github.com/mlposey/z4/feeds"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type Config struct {
	DB   *storage.BadgerClient
	Port int
	Opts []grpc.ServerOption
}

type Server struct {
	fm     *feeds.Manager
	config Config
	server *grpc.Server
}

func NewServer(config Config) *Server {
	return &Server{config: config}
}

func (s *Server) Start() error {
	telemetry.Logger.Info("starting server...")
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.config.Port))
	if err != nil {
		return err
	}
	telemetry.Logger.Info("listening for connections",
		zap.Int("port", s.config.Port))

	s.server = grpc.NewServer(s.config.Opts...)
	proto.RegisterAdminServer(s.server, &admin{})

	s.fm = feeds.NewManager(s.config.DB)
	proto.RegisterCollectionServer(s.server, newCollection(s.fm))
	return s.server.Serve(lis)
}

func (s *Server) Close() error {
	// TODO: Consider supporting a context with timeout.

	telemetry.Logger.Info("stopping server...")
	s.server.GracefulStop()
	err := s.fm.Close()
	telemetry.Logger.Info("server stopped")
	return err
}
