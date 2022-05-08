package sched

import (
	"context"
	"github.com/mlposey/z4/feeds/q"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"time"
)

type taskReader struct {
	namespace *proto.Namespace
	tasks     *storage.TaskStore
	ctx       context.Context
	pipe      chan *proto.Task
}

func NewScheduledTaskReader(
	ctx context.Context,
	namespace *proto.Namespace,
	tasks *storage.TaskStore,
) *taskReader {
	reader := &taskReader{
		namespace: namespace,
		tasks:     tasks,
		pipe:      make(chan *proto.Task),
		ctx:       ctx,
	}
	go reader.startReadLoop()
	return reader
}

func (tr *taskReader) Tasks() q.TaskStream {
	return tr.pipe
}

func (tr *taskReader) startReadLoop() {
	for {
		select {
		case <-tr.ctx.Done():
			telemetry.Logger.Info("scheduled task reader stopped",
				zap.String("namespace", tr.namespace.GetId()))
			return
		default:
		}

		pushCount, err := tr.pullAndPush()
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
func (tr *taskReader) pullAndPush() (int, error) {
	lastDeliveredTaskID := tr.namespace.LastTask
	dc, err1 := tr.processDelivered(lastDeliveredTaskID)
	uc, err2 := tr.processUndelivered(lastDeliveredTaskID)
	return dc + uc, multierr.Combine(err1, err2)
}

func (tr *taskReader) push(task *proto.Task) error {
	select {
	case <-tr.ctx.Done():
		return io.EOF

	case tr.pipe <- task:
		return nil
	}
}

// attempts to redeliver unacknowledged tasks
func (tr *taskReader) processDelivered(lastID string) (int, error) {
	ackDeadline := time.Second * time.Duration(tr.namespace.GetAckDeadlineSeconds())
	df := &deliveredTaskFetcher{
		Tasks:           tr.tasks,
		LastDeliveredID: lastID,
		Namespace:       tr.namespace.GetId(),
		AckDeadline:     ackDeadline,
	}

	var tasks []*proto.Task
	err := df.Process(func(task *proto.Task) error {
		if err := tr.push(task); err != nil {
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
	err = tr.tasks.SaveAll(tasks)
	if err != nil {
		telemetry.Logger.Error("failed to update retry tasks",
			zap.Error(err))
		return len(tasks), err
	}
	return len(tasks), nil
}

// attempts to deliver new tasks
func (tr *taskReader) processUndelivered(lastID string) (int, error) {
	uf := &undeliveredTaskFetcher{
		Tasks:     tr.tasks,
		StartID:   lastID,
		Namespace: tr.namespace.GetId(),
	}

	var count int
	err := uf.Process(func(task *proto.Task) error {
		if err := tr.push(task); err != nil {
			return err
		}
		tr.namespace.LastTask = task.GetId()
		count++
		return nil
	})

	if err != nil && err != io.EOF {
		return count, err
	}
	return count, nil
}
