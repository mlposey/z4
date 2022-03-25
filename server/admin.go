package server

import (
	"context"
	"z4/proto"
)

type admin struct {
	proto.UnimplementedAdminServer
}

func (a *admin) CheckHealth(ctx context.Context, req *proto.CheckHealthRequest) (*proto.Status, error) {
	// TODO: Implement real health check.
	return new(proto.Status), nil
}
