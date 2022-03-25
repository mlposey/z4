package server

import (
	"fmt"
	"google.golang.org/grpc"
	"net"
	"z4/proto"
)

// Start opens the port to incoming gRPC requests.
func Start(port int, opts []grpc.ServerOption) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return err
	}

	server := grpc.NewServer(opts...)
	proto.RegisterAdminServer(server, &admin{})
	proto.RegisterCollectionServer(server, &collection{})
	return server.Serve(lis)
}
