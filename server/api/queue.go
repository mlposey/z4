package api

import (
	"context"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/feeds"
	"github.com/mlposey/z4/iden"
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
	ids    *storage.IDGenerator
}

func NewQueue(
	fm *feeds.Manager,
	tasks *storage.TaskStore,
	raft *raft.Raft,
	handle *cluster.LeaderHandle,
	ids *storage.IDGenerator,
) proto.QueueServer {
	return &Queue{
		fm:     fm,
		tasks:  tasks,
		raft:   raft,
		handle: handle,
		ids:    ids,
	}
}

func (q *Queue) Push(ctx context.Context, req *proto.PushTaskRequest) (*proto.PushTaskResponse, error) {
	telemetry.PushedTasks.
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
		telemetry.PushedTasks.
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
		res, err := q.forwardPushRequest(ctx, req)
		if err == nil {
			leader := q.handle.LeaderAddress()
			// don't overwrite if it was forwarded multiple times
			if res.GetForwardedTo() == "" {
				res.ForwardedTo = leader
			}
		}
		return res, err
	}

	task, err := q.makeTask(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "task creation failed: %v", err)
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

func (q *Queue) forwardPushRequest(
	ctx context.Context,
	req *proto.PushTaskRequest,
) (*proto.PushTaskResponse, error) {

	client, err := q.handle.QueueClient()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not forward request: %v", err)
	}
	return client.Push(ctx, req)
}

func (q *Queue) makeTask(req *proto.PushTaskRequest) (*proto.Task, error) {
	task := &proto.Task{
		Namespace: req.GetNamespace(),
		Metadata:  req.GetMetadata(),
		Payload:   req.GetPayload(),
		CreatedAt: timestamppb.New(time.Now()),
	}

	ts := q.getRunTime(req)
	id, err := q.ids.ID(req.GetNamespace(), ts)
	if err != nil {
		return nil, err
	}

	if ts.IsZero() {
		task.Id = id.String()
	} else {
		task.Id = id.String()
		task.ScheduleTime = timestamppb.New(ts)
	}
	return task, nil
}

func (q *Queue) getRunTime(req *proto.PushTaskRequest) time.Time {
	if req.GetTtsSeconds() > 0 {
		return time.Now().Add(time.Duration(req.GetTtsSeconds()) * time.Second)
	} else if req.GetScheduleTime() != nil {
		return req.GetScheduleTime().AsTime()
	}
	return time.Time{}
}

func (q *Queue) GetTask(ctx context.Context, req *proto.GetTaskRequest) (*proto.Task, error) {
	id, err := iden.ParseString(req.GetReference().GetTaskId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid task id")
	}

	task, err := q.tasks.Get(req.GetReference().GetNamespace(), id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "task not found: %v", err)
	}

	telemetry.PulledTasks.
		WithLabelValues("GetTask", req.GetReference().GetNamespace()).
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

func (q *Queue) Delete(ctx context.Context, req *proto.DeleteTaskRequest) (*proto.DeleteTaskResponse, error) {
	telemetry.RemovedTasks.
		WithLabelValues("Delete", req.GetReference().GetNamespace()).
		Inc()

	if !q.handle.IsLeader() {
		res, err := q.forwardDeleteRequest(ctx, req)
		if err == nil {
			leader := q.handle.LeaderAddress()
			// don't overwrite if it was forwarded multiple times
			if res.GetForwardedTo() == "" {
				res.ForwardedTo = leader
			}
		}
		return res, err
	}

	ack := &proto.Ack{
		Reference: req.GetReference(),
	}

	if req.GetAsync() {
		cluster.ApplyAckCommand(q.raft, ack)
	} else {
		err := cluster.ApplyAckCommand(q.raft, ack).Error()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to delete task: %v", err)
		}
	}
	return new(proto.DeleteTaskResponse), nil
}

func (q *Queue) forwardDeleteRequest(
	ctx context.Context,
	req *proto.DeleteTaskRequest,
) (*proto.DeleteTaskResponse, error) {

	client, err := q.handle.QueueClient()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not forward request: %v", err)
	}
	return client.Delete(ctx, req)
}
