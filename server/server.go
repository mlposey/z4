package server

import (
	"fmt"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

// Start opens the port to incoming gRPC requests.
func Start(port int, opts []grpc.ServerOption) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return err
	}
	telemetry.Logger.Info("listening for connections",
		zap.Int("port", port))

	server := grpc.NewServer(opts...)
	proto.RegisterAdminServer(server, &admin{})
	proto.RegisterCollectionServer(server, newCollection())
	return server.Serve(lis)
}
