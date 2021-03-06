package feeds

import (
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/server/cluster/sm"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TaskBroker manages the read and ack streams of a task feed.
type TaskBroker struct {
	fm     *Manager
	stream proto.Queue_PullServer
	raft   *raft.Raft
	queue  string
	lease  *Lease
}

func NewTaskBroker(
	stream proto.Queue_PullServer,
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
	md, ok := metadata.FromIncomingContext(tb.stream.Context())
	if !ok {
		return status.Errorf(codes.InvalidArgument, "missing metadata")
	}
	ns := md.Get("queue")
	if len(ns) != 1 {
		return status.Errorf(codes.InvalidArgument, "expected one queue in metadata")
	}
	tb.queue = ns[0]

	l, err := tb.fm.Lease(tb.queue)
	if err != nil {
		return status.Errorf(codes.Internal, "could not start task feed: %v", err)
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
		ack, err := tb.stream.Recv()
		if err != nil {
			telemetry.Logger.Debug("closing ack stream",
				zap.Error(err),
				zap.String("queue", tb.queue))
			return
		}

		telemetry.RemovedTasks.
			WithLabelValues("Ack", ack.GetReference().GetQueue()).
			Inc()

		// This is intentionally async.
		sm.ApplyAckCommand(tb.raft, ack)
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

			telemetry.PulledTasks.
				WithLabelValues("Pull", tb.queue).
				Inc()
		}
	}
}
