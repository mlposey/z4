package server

import (
	"context"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/feeds"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"time"
)

// collection implements the gRPC Collection service.
type collection struct {
	proto.UnimplementedCollectionServer
	fm   *feeds.Manager
	raft *raft.Raft
}

func newCollection(fm *feeds.Manager, raft *raft.Raft) proto.CollectionServer {
	return &collection{fm: fm, raft: raft}
}

type taskCreationType int

const (
	asyncCreation taskCreationType = iota
	syncCreation
)

func (c *collection) CreateTask(ctx context.Context, req *proto.CreateTaskRequest) (*proto.Task, error) {
	telemetry.Logger.Debug("got CreateTask rpc request")
	return c.createTask(ctx, req, syncCreation)
}

func (c *collection) CreateTaskAsync(ctx context.Context, req *proto.CreateTaskRequest) (*proto.Task, error) {
	telemetry.Logger.Debug("got CreateTaskAsync rpc request")
	return c.createTask(ctx, req, asyncCreation)
}

func (c *collection) CreateTaskStream(stream proto.Collection_CreateTaskStreamServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			telemetry.Logger.Info("client closed CreateTaskStreamAsyncV2 stream")
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to get tasks from client: %v", err)
		}

		task, err := c.createTask(stream.Context(), req, syncCreation)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to create task: %v", err)
		}

		err = stream.Send(&proto.TaskStreamResponse{
			Task:   task,
			Status: uint32(codes.OK),
		})

		if err != nil {
			telemetry.Logger.Error("failed to send task to client", zap.Error(err))
			return status.Errorf(codes.Internal, "failed to send task to client")
		}
	}
}

func (c *collection) CreateTaskStreamAsync(stream proto.Collection_CreateTaskStreamAsyncServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			telemetry.Logger.Info("client closed CreateTaskStreamAsyncV2 stream")
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to get tasks from client: %v", err)
		}

		task, err := c.createTask(stream.Context(), req, asyncCreation)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to create task: %v", err)
		}

		err = stream.Send(&proto.TaskStreamResponse{
			Task:   task,
			Status: uint32(codes.OK),
		})

		if err != nil {
			telemetry.Logger.Error("failed to send task to client", zap.Error(err))
			return status.Errorf(codes.Internal, "failed to send task to client")
		}
	}
}

func (c *collection) createTask(ctx context.Context, req *proto.CreateTaskRequest, ct taskCreationType) (*proto.Task, error) {
	task := &proto.Task{
		Id:        storage.NewTaskID(c.getRunTime(req)),
		Namespace: req.GetNamespace(),
		DeliverAt: timestamppb.New(c.getRunTime(req)),
		Metadata:  req.GetMetadata(),
		Payload:   req.GetPayload(),
	}

	switch ct {
	case asyncCreation:
		c.saveTask(task)

	case syncCreation:
		err := c.saveTask(task).Error()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to save task: %v", err)
		}

	default:
		return nil, status.Errorf(codes.Internal, "invalid creation type %v", ct)
	}

	return task, nil
}

func (c *collection) saveTask(task *proto.Task) raft.ApplyFuture {
	cmd, _ := pb.Marshal(task)
	return c.raft.Apply(cmd, 0)
}

func (c *collection) GetTaskStream(req *proto.StreamTasksRequest, stream proto.Collection_GetTaskStreamServer) error {
	telemetry.Logger.Debug("got StreamTasks rpc request")
	lease := c.fm.Lease(req.GetNamespace())
	defer lease.Release()

	tasks := lease.Feed().Tasks()
	for task := range tasks {
		telemetry.Logger.Debug("sending task to client",
			zap.Any("task", task),
			telemetry.LogRequestID(req.GetRequestId()))

		err := stream.Send(task)

		if err != nil {
			return status.Errorf(codes.Internal, "failed to send tasks to client: %v", err)
		}
	}

	telemetry.Logger.Info("client closed GetTaskStream stream",
		telemetry.LogRequestID(req.GetRequestId()))
	return nil
}

func (c *collection) getRunTime(req *proto.CreateTaskRequest) time.Time {
	if req.GetTtsSeconds() > 0 {
		return time.Now().Add(time.Duration(req.GetTtsSeconds()) * time.Second)
	}
	return req.GetDeliverAt().AsTime()
}
