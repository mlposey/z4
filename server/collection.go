package server

import (
	"context"
	"github.com/mlposey/z4/feeds"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

// collection implements the gRPC Collection service.
type collection struct {
	proto.UnimplementedCollectionServer
	fm *feeds.Manager
}

func newCollection(fm *feeds.Manager) proto.CollectionServer {
	return &collection{fm: fm}
}

func (c *collection) CreateTask(ctx context.Context, req *proto.CreateTaskRequest) (*proto.Task, error) {
	telemetry.Logger.Debug("got CreateTask rpc request")
	lease := c.fm.Lease(req.GetNamespace())
	defer lease.Release()

	task := storage.Task{
		ID:        storage.NewTaskID(c.getRunTime(req)),
		Namespace: req.GetNamespace(),
		RunTime:   c.getRunTime(req),
		Metadata:  req.GetMetadata(),
		Payload:   req.GetPayload(),
	}
	// TODO: Create separate endpoint for async creation.
	/*err := lease.Feed().Add(task)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save storage: %v", err)
	}*/
	lease.Feed().AddAsync(task)
	return &proto.Task{
		Metadata:  task.Metadata,
		Payload:   task.Payload,
		DeliverAt: timestamppb.New(task.RunTime),
		Id:        task.ID,
		Namespace: task.Namespace,
	}, nil
}

func (c *collection) StreamTasks(req *proto.StreamTasksRequest, stream proto.Collection_StreamTasksServer) error {
	telemetry.Logger.Debug("got StreamTasks rpc request")
	lease := c.fm.Lease(req.GetNamespace())
	defer lease.Release()

	tasks := lease.Feed().Tasks()
	for task := range tasks {
		telemetry.Logger.Debug("sending task to client",
			zap.Any("task", task),
			telemetry.LogRequestID(req.GetRequestId()))

		err := stream.Send(&proto.Task{
			Metadata:  task.Metadata,
			Payload:   task.Payload,
			DeliverAt: timestamppb.New(task.RunTime),
			Id:        task.ID,
			Namespace: task.Namespace,
		})

		if err != nil {
			return status.Errorf(codes.Internal, "failed to send tasks to client: %v", err)
		}
	}

	telemetry.Logger.Info("client closed storage stream",
		telemetry.LogRequestID(req.GetRequestId()))
	return nil
}

func (c *collection) getRunTime(req *proto.CreateTaskRequest) time.Time {
	if req.GetTtsSeconds() > 0 {
		return time.Now().Add(time.Duration(req.GetTtsSeconds()) * time.Second)
	}
	return req.GetDeliverAt().AsTime()
}
