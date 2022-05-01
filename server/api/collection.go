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
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"time"
)

// Queue implements the gRPC Queue service.
type Queue struct {
	proto.UnimplementedQueueServer
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
) proto.QueueServer {
	return &Queue{
		fm:     fm,
		tasks:  tasks,
		raft:   raft,
		handle: handle,
	}
}

func (q *Queue) Push(ctx context.Context, req *proto.PushTaskRequest) (*proto.PushTaskResponse, error) {
	telemetry.PushTaskRequests.
		WithLabelValues("Push", req.GetNamespace()).
		Inc()
	return q.createTask(ctx, req)
}

func (q *Queue) PushStream(stream proto.Queue_PushStreamServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to get tasks from client: %v", err)
		}
		telemetry.PushTaskRequests.
			WithLabelValues("PushStream", req.GetNamespace()).
			Inc()

		task, err := q.createTask(stream.Context(), req)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to create task: %v", err)
		}

		err = stream.Send(&proto.PushStreamResponse{
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

func (q *Queue) createTask(ctx context.Context, req *proto.PushTaskRequest) (*proto.PushTaskResponse, error) {
	if !q.handle.IsLeader() {
		leader := q.handle.LeaderAddress()
		res, err := q.forwardRequest(ctx, req)
		if err == nil {
			// don't overwrite if it was forwarded multiple times
			if res.GetForwardedTo() == "" {
				res.ForwardedTo = leader
			}
		}
		return res, err
	}

	task := &proto.Task{
		Id:        storage.NewTaskID(q.getRunTime(req)),
		Namespace: req.GetNamespace(),
		DeliverAt: timestamppb.New(q.getRunTime(req)),
		Metadata:  req.GetMetadata(),
		Payload:   req.GetPayload(),
	}

	if req.GetAsync() {
		cluster.ApplySaveTaskCommand(q.raft, task)
	} else {
		err := cluster.ApplySaveTaskCommand(q.raft, task).Error()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to save task: %v", err)
		}
	}
	return &proto.PushTaskResponse{Task: task}, nil
}

func (q *Queue) forwardRequest(
	ctx context.Context,
	req *proto.PushTaskRequest,
) (*proto.PushTaskResponse, error) {

	client, err := q.handle.QueueClient()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not forward request: %v", err)
	}
	return client.Push(ctx, req)
}

func (q *Queue) GetTask(ctx context.Context, req *proto.GetTaskRequest) (*proto.Task, error) {
	task, err := q.tasks.Get(req.GetNamespace(), req.GetTaskId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "task not found: %v", err)
	}

	telemetry.StreamedTasks.
		WithLabelValues("GetTask", req.GetNamespace()).
		Inc()
	return task, nil
}

func (q *Queue) Pull(stream proto.Queue_PullServer) error {
	if !q.handle.IsLeader() {
		return status.Errorf(codes.FailedPrecondition, "rpc must be called on cluster leader")
	}

	broker := feeds.NewTaskBroker(stream, q.fm, q.raft)
	defer broker.Close()
	return broker.Start()
}

func (q *Queue) getRunTime(req *proto.PushTaskRequest) time.Time {
	if req.GetTtsSeconds() > 0 {
		return time.Now().Add(time.Duration(req.GetTtsSeconds()) * time.Second)
	}
	return req.GetDeliverAt().AsTime()
}
