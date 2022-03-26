package server

import (
	"fmt"
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

// Start opens the port to incoming gRPC requests.
func Start(config Config) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.Port))
	if err != nil {
		return err
	}
	telemetry.Logger.Info("listening for connections",
		zap.Int("port", config.Port))

	server := grpc.NewServer(config.Opts...)
	proto.RegisterAdminServer(server, &admin{})
	proto.RegisterCollectionServer(server, newCollection(config.DB))
	return server.Serve(lis)
}
