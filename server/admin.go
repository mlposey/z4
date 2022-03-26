package server

import (
	"context"
	"github.com/mlposey/z4/proto"
)

// admin implements the gRPC Admin service.
type admin struct {
	proto.UnimplementedAdminServer
}

func (a *admin) CheckHealth(ctx context.Context, req *proto.CheckHealthRequest) (*proto.Status, error) {
	// TODO: Implement real health check.
	return new(proto.Status), nil
}
