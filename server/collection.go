package server

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"z4/proto"
	"z4/storage"
)

type collection struct {
	proto.UnimplementedCollectionServer
	store *storage.SimpleStore
}

func newCollection() proto.CollectionServer {
	return &collection{store: new(storage.SimpleStore)}
}

func (c *collection) CreateTask(ctx context.Context, req *proto.CreateTaskRequest) (*proto.Task, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTask not implemented")
}

func (c *collection) StreamTasks(req *proto.StreamTasksRequest, stream proto.Collection_StreamTasksServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamTasks not implemented")
}
