package feeds

import (
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/server/cluster"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TaskBroker manages the read and ack streams of a task feed.
type TaskBroker struct {
	fm        *Manager
	stream    proto.Collection_GetTaskStreamServer
	raft      *raft.Raft
	namespace string
	lease     *Lease
}

func NewTaskBroker(
	stream proto.Collection_GetTaskStreamServer,
	fm *Manager,
	raft *raft.Raft,
) *TaskBroker {
	return &TaskBroker{
		fm:     fm,
		stream: stream,
		raft:   raft,
	}
}

func (tb *TaskBroker) Start() error {
	start, err := tb.stream.Recv()
	if err != nil {
		return status.Errorf(codes.Internal, "failed to start GetTaskStream stream")
	}

	if start.GetStartReq() == nil {
		return status.Error(codes.InvalidArgument, "got unexpected start request")
	}

	tb.namespace = start.GetStartReq().GetNamespace()
	l, err := tb.fm.Lease(tb.namespace)
	if err != nil {
		return status.Errorf(codes.Internal, "could not start task feed")
	}
	tb.lease = l

	go tb.startAckListener()
	return tb.startTaskSender()
}

func (tb *TaskBroker) Close() error {
	if tb.lease != nil {
		tb.lease.Release()
	}
	return nil
}

func (tb *TaskBroker) startAckListener() {
	for {
		req, err := tb.stream.Recv()
		if err != nil {
			telemetry.Logger.Debug("closing ack stream",
				zap.Error(err),
				zap.String("namespace", tb.namespace))
			return
		}

		if req.GetAck() == nil {
			telemetry.Logger.Warn("got invalid ack",
				zap.String("namespace", tb.namespace))
			continue
		}

		telemetry.Logger.Debug("got ack",
			zap.String("namespace", tb.namespace),
			zap.String("task_id", req.GetAck().GetTaskId()))

		cluster.ApplyAckCommand(tb.raft, req.GetAck())
	}
}

func (tb *TaskBroker) startTaskSender() error {
	for {
		select {
		case <-tb.stream.Context().Done():
			return nil

		case task := <-tb.lease.Feed().Tasks():
			err := tb.stream.Send(task)
			if err != nil {
				return status.Errorf(codes.Internal, "failed to send tasks to client: %v", err)
			}

			telemetry.StreamedTasks.
				WithLabelValues("GetTaskStream", tb.namespace).
				Inc()
		}
	}
}
