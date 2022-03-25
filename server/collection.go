package server

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	t, err := c.store.Save(ctx, storage.TaskDefinition{
		RunTime:  req.GetDeliverAt().AsTime(),
		Metadata: req.GetMetadata(),
		Payload:  req.GetPayload(),
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save task: %v", err)
	}
	return &proto.Task{
		Metadata:  t.Metadata,
		Payload:   t.Payload,
		DeliverAt: timestamppb.New(t.RunTime),
		Id:        t.ID,
	}, nil
}

func (c *collection) StreamTasks(req *proto.StreamTasksRequest, stream proto.Collection_StreamTasksServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamTasks not implemented")
}
