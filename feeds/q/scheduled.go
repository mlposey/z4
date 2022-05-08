package q

import (
	"context"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"time"
)

type scheduledTaskReader struct {
	namespace *proto.Namespace
	tasks     *storage.TaskStore
	ctx       context.Context
	pipe      chan *proto.Task
}

func NewScheduledTaskReader(
	ctx context.Context,
	namespace *proto.Namespace,
	tasks *storage.TaskStore,
) *scheduledTaskReader {
	reader := &scheduledTaskReader{
		namespace: namespace,
		tasks:     tasks,
		pipe:      make(chan *proto.Task),
		ctx:       ctx,
	}
	go reader.startReadLoop()
	return reader
}

func (q *scheduledTaskReader) Tasks() TaskStream {
	return q.pipe
}

func (q *scheduledTaskReader) startReadLoop() {
	for {
		select {
		case <-q.ctx.Done():
			telemetry.Logger.Info("scheduled task reader stopped",
				zap.String("namespace", q.namespace.GetId()))
			return
		default:
		}

		pushCount, err := q.pullAndPush()
		if err != nil {
			telemetry.Logger.Error("feed operation failed", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		if pushCount == 0 {
			time.Sleep(time.Millisecond * 50)
			continue
		}
	}
}

// pullAndPush loads ready tasks from storage and delivers them to consumers.
func (q *scheduledTaskReader) pullAndPush() (int, error) {
	lastDeliveredTaskID := q.namespace.LastDeliveredTask
	dc, err1 := q.processDelivered(lastDeliveredTaskID)
	uc, err2 := q.processUndelivered(lastDeliveredTaskID)
	return dc + uc, multierr.Combine(err1, err2)
}

func (q *scheduledTaskReader) push(task *proto.Task) error {
	select {
	case <-q.ctx.Done():
		return io.EOF

	case q.pipe <- task:
		return nil
	}
}

// attempts to redeliver unacknowledged tasks
func (q *scheduledTaskReader) processDelivered(lastID string) (int, error) {
	ackDeadline := time.Second * time.Duration(q.namespace.GetAckDeadlineSeconds())
	df := &deliveredTaskFetcher{
		Tasks:           q.tasks,
		LastDeliveredID: lastID,
		Namespace:       q.namespace.GetId(),
		AckDeadline:     ackDeadline,
	}

	var tasks []*proto.Task
	err := df.Process(func(task *proto.Task) error {
		if err := q.push(task); err != nil {
			return err
		}

		task.LastRetry = timestamppb.New(time.Now())
		tasks = append(tasks, task)
		return nil
	})
	if err != nil && err != io.EOF {
		return len(tasks), err
	}

	if len(tasks) == 0 {
		return 0, nil
	}

	// TODO: Determine if we want to apply this to the Raft log.
	// It doesn't seem entirely necessary, and not having it should
	// improve performance. It could be nice to have though.
	// TODO: Batch task writes using buffers instead of saving one large slice.
	err = q.tasks.SaveAll(tasks)
	if err != nil {
		telemetry.Logger.Error("failed to update retry tasks",
			zap.Error(err))
		return len(tasks), err
	}
	return len(tasks), nil
}

// attempts to deliver new tasks
func (q *scheduledTaskReader) processUndelivered(lastID string) (int, error) {
	uf := &undeliveredTaskFetcher{
		Tasks:     q.tasks,
		StartID:   lastID,
		Namespace: q.namespace.GetId(),
	}

	var count int
	err := uf.Process(func(task *proto.Task) error {
		if err := q.push(task); err != nil {
			return err
		}
		q.namespace.LastDeliveredTask = task.GetId()
		count++
		return nil
	})

	if err != nil && err != io.EOF {
		return count, err
	}
	return count, nil
}
