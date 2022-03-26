package server

import (
	"context"
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
	qm *storage.QueueManager
}

func newCollection(db *storage.BadgerClient) proto.CollectionServer {
	return &collection{
		qm: storage.NewQueueManager(db),
	}
}

func (c *collection) CreateTask(ctx context.Context, req *proto.CreateTaskRequest) (*proto.Task, error) {
	lease := c.qm.Lease(req.GetNamespace())
	defer lease.Release()

	task := storage.NewTask(storage.TaskDefinition{
		//RunTime:  req.GetDeliverAt().AsTime(),
		RunTime:  time.Now().Add(time.Duration(req.GetTtsSeconds()) * time.Second),
		Metadata: req.GetMetadata(),
		Payload:  req.GetPayload(),
	})
	err := lease.Queue().Add(task)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save storage: %v", err)
	}
	return &proto.Task{
		Metadata:  task.Metadata,
		Payload:   task.Payload,
		DeliverAt: timestamppb.New(task.RunTime),
		Id:        task.ID.String(),
	}, nil
}

func (c *collection) StreamTasks(req *proto.StreamTasksRequest, stream proto.Collection_StreamTasksServer) error {
	lease := c.qm.Lease(req.GetNamespace())
	defer lease.Release()

	tasks := lease.Queue().Feed()
	for task := range tasks {
		telemetry.Logger.Debug("sending task to client",
			zap.Any("task", task),
			telemetry.LogRequestID(req.GetRequestId()))

		err := stream.Send(&proto.Task{
			Metadata:  task.Metadata,
			Payload:   task.Payload,
			DeliverAt: timestamppb.New(task.RunTime),
			Id:        task.ID.String(),
		})

		if err != nil {
			return status.Errorf(codes.Internal, "failed to send tasks to client: %v", err)
		}
	}

	telemetry.Logger.Info("client closed storage stream",
		telemetry.LogRequestID(req.GetRequestId()))
	return nil
}
