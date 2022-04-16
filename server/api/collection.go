package api

import (
	"context"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/feeds"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/server/cluster"
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

// Collection implements the gRPC Collection service.
type Collection struct {
	proto.UnimplementedCollectionServer
	fm     *feeds.Manager
	tasks  *storage.TaskStore
	raft   *raft.Raft
	handle *cluster.LeaderHandle
}

func NewCollection(
	fm *feeds.Manager,
	tasks *storage.TaskStore,
	raft *raft.Raft,
	handle *cluster.LeaderHandle,
) proto.CollectionServer {
	return &Collection{
		fm:     fm,
		tasks:  tasks,
		raft:   raft,
		handle: handle,
	}
}

type taskCreationType int

const (
	asyncCreation taskCreationType = iota
	syncCreation
)

func (c *Collection) CreateTask(ctx context.Context, req *proto.CreateTaskRequest) (*proto.CreateTaskResponse, error) {
	telemetry.CreateTaskRequests.
		WithLabelValues("CreateTask", req.GetNamespace()).
		Inc()
	return c.createTask(ctx, req, syncCreation)
}

func (c *Collection) CreateTaskAsync(ctx context.Context, req *proto.CreateTaskRequest) (*proto.CreateTaskResponse, error) {
	telemetry.CreateTaskRequests.
		WithLabelValues("CreateTaskAsync", req.GetNamespace()).
		Inc()
	return c.createTask(ctx, req, asyncCreation)
}

func (c *Collection) CreateTaskStream(stream proto.Collection_CreateTaskStreamServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to get tasks from client: %v", err)
		}
		telemetry.CreateTaskRequests.
			WithLabelValues("CreateTaskStream", req.GetNamespace()).
			Inc()

		task, err := c.createTask(stream.Context(), req, syncCreation)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to create task: %v", err)
		}

		err = stream.Send(&proto.TaskStreamResponse{
			Task:        task.GetTask(),
			Status:      uint32(codes.OK),
			ForwardedTo: task.GetForwardedTo(),
		})

		if err != nil {
			telemetry.Logger.Error("failed to send task to client", zap.Error(err))
			return status.Errorf(codes.Internal, "failed to send task to client")
		}
	}
}

func (c *Collection) CreateTaskStreamAsync(stream proto.Collection_CreateTaskStreamAsyncServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to get tasks from client: %v", err)
		}
		telemetry.CreateTaskRequests.
			WithLabelValues("CreateTaskStreamAsync", req.GetNamespace()).
			Inc()

		task, err := c.createTask(stream.Context(), req, asyncCreation)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to create task: %v", err)
		}

		err = stream.Send(&proto.TaskStreamResponse{
			Task:        task.GetTask(),
			Status:      uint32(codes.OK),
			ForwardedTo: task.GetForwardedTo(),
		})

		if err != nil {
			telemetry.Logger.Error("failed to send task to client", zap.Error(err))
			return status.Errorf(codes.Internal, "failed to send task to client")
		}
	}
}

func (c *Collection) createTask(ctx context.Context, req *proto.CreateTaskRequest, ct taskCreationType) (*proto.CreateTaskResponse, error) {
	if !c.handle.IsLeader() {
		leader := c.handle.LeaderAddress()
		res, err := c.forwardRequest(ctx, req, ct)
		if err == nil {
			// don't overwrite if it was forwarded multiple times
			if res.GetForwardedTo() == "" {
				res.ForwardedTo = leader
			}
		}
		return res, err
	}

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

	return &proto.CreateTaskResponse{Task: task}, nil
}

func (c *Collection) forwardRequest(
	ctx context.Context,
	req *proto.CreateTaskRequest,
	ct taskCreationType,
) (*proto.CreateTaskResponse, error) {

	client, err := c.handle.CollectionClient()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not forward request: %v", err)
	}

	switch ct {
	case asyncCreation:
		return client.CreateTaskAsync(ctx, req)

	case syncCreation:
		return client.CreateTask(ctx, req)

	default:
		return nil, status.Errorf(codes.Internal, "invalid creation type %v", ct)
	}
}

func (c *Collection) saveTask(task *proto.Task) raft.ApplyFuture {
	cmd, _ := pb.Marshal(task)
	return c.raft.Apply(cmd, 0)
}

func (c *Collection) GetTask(ctx context.Context, req *proto.GetTaskRequest) (*proto.Task, error) {
	task, err := c.tasks.Get(req.GetNamespace(), req.GetTaskId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "task not found: %v", err)
	}

	telemetry.StreamedTasks.
		WithLabelValues("GetTask", req.GetNamespace()).
		Inc()
	return task, nil
}

func (c *Collection) GetTaskStream(req *proto.StreamTasksRequest, stream proto.Collection_GetTaskStreamServer) error {
	namespace := req.GetNamespace()
	return c.fm.Tasks(namespace, func(tasks feeds.TaskStream) error {
		for {
			select {
			case <-stream.Context().Done():
				return nil

			case task := <-tasks:
				err := stream.Send(task)
				if err != nil {
					return status.Errorf(codes.Internal, "failed to send tasks to client: %v", err)
				}

				telemetry.StreamedTasks.
					WithLabelValues("GetTaskStream", namespace).
					Inc()
			}
		}
	})
}

func (c *Collection) getRunTime(req *proto.CreateTaskRequest) time.Time {
	if req.GetTtsSeconds() > 0 {
		return time.Now().Add(time.Duration(req.GetTtsSeconds()) * time.Second)
	}
	return req.GetDeliverAt().AsTime()
}
