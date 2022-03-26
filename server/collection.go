package server

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"z4/proto"
	"z4/queue"
	"z4/storage"
	"z4/telemetry"
)

type collection struct {
	proto.UnimplementedCollectionServer
	tasks *queue.Tasks
}

func newCollection() proto.CollectionServer {
	return &collection{
		tasks: queue.New(),
	}
}

func (c *collection) CreateTask(ctx context.Context, req *proto.CreateTaskRequest) (*proto.Task, error) {
	t, err := c.tasks.Add(ctx, storage.TaskDefinition{
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
	tasks := c.tasks.Feed(stream.Context())
	for task := range tasks {
		telemetry.Logger.Debug("sending task to client",
			zap.Any("task", task),
			telemetry.LogRequestID(req.GetRequestId()))

		err := stream.Send(&proto.Task{
			Metadata:  task.Metadata,
			Payload:   task.Payload,
			DeliverAt: timestamppb.New(task.RunTime),
			Id:        task.ID,
		})

		if err != nil {
			return status.Errorf(codes.Internal, "failed to send tasks to client: %v", err)
		}
	}

	telemetry.Logger.Info("client closed task stream",
		telemetry.LogRequestID(req.GetRequestId()))
	return nil
}
