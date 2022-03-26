package server

import (
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"z4/proto"
	"z4/telemetry"
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
