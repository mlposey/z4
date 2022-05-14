package sched

import (
	"context"
	"github.com/mlposey/z4/feeds/q"
	"github.com/mlposey/z4/iden"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

type taskReader struct {
	namespace *proto.Namespace
	tasks     *storage.TaskStore
	ctx       context.Context
	pipe      chan *proto.Task
	sweepTag  uint64
	sweep     sync.Mutex
	lastSweep time.Time
}

func NewScheduledTaskReader(
	ctx context.Context,
	namespace *proto.Namespace,
	tasks *storage.TaskStore,
) *taskReader {
	reader := &taskReader{
		namespace: namespace,
		tasks:     tasks,
		pipe:      make(chan *proto.Task, 100_000),
		ctx:       ctx,
	}
	go reader.startReadLoop()
	return reader
}

func (tr *taskReader) Tasks() q.TaskStream {
	return tr.pipe
}

func (tr *taskReader) startReadLoop() {
	prefetch := 1_000
	for {
		select {
		case <-tr.ctx.Done():
			telemetry.Logger.Info("scheduled task reader stopped",
				zap.String("namespace", tr.namespace.GetId()))
			return
		default:
		}

		pushCount, err := tr.pullAndPush(prefetch)
		if err != nil {
			telemetry.Logger.Error("feed operation failed", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		// TODO: Intelligently pick prefetch size based on push count.
		if pushCount == 0 {
			time.Sleep(time.Millisecond * 500)
			prefetch = 0
		} else {
			prefetch = 1_000
		}
	}
}

// pullAndPush loads ready tasks from storage and delivers them to consumers.
func (tr *taskReader) pullAndPush(prefetch int) (int, error) {
	lastID := iden.MustParseString(tr.namespace.LastDeliveredScheduledTask)
	if time.Since(tr.lastSweep) > time.Second*30 {
		tr.lastSweep = time.Now()
		go func(tag uint64) {
			tr.sweep.Lock()
			defer tr.sweep.Unlock()
			if !atomic.CompareAndSwapUint64(&tr.sweepTag, tag, tag+1) {
				return
			}

			err := tr.processDelivered(lastID)
			if err != nil {
				telemetry.Logger.Error("failed to process delivered scheduled tasks",
					zap.Error(err))
			}
		}(atomic.LoadUint64(&tr.sweepTag))
	}
	return tr.processUndelivered(lastID, prefetch)
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
func (tr *taskReader) processDelivered(lastID iden.TaskID) error {
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
		return err
	}

	if len(tasks) == 0 {
		return nil
	}

	// TODO: Batch task writes using buffers instead of saving one large slice.
	err = tr.tasks.SaveAll(tasks)
	if err != nil {
		telemetry.Logger.Error("failed to update retry tasks",
			zap.Error(err))
		return err
	}
	return nil
}

// attempts to deliver new tasks
func (tr *taskReader) processUndelivered(lastID iden.TaskID, prefetch int) (int, error) {
	uf := &undeliveredTaskFetcher{
		Tasks:     tr.tasks,
		StartID:   lastID,
		Namespace: tr.namespace.GetId(),
		Prefetch:  prefetch,
	}

	var count int
	err := uf.Process(func(task *proto.Task) error {
		if err := tr.push(task); err != nil {
			return err
		}
		tr.namespace.LastDeliveredScheduledTask = task.GetId()
		count++
		return nil
	})

	if err != nil && err != io.EOF {
		return count, err
	}
	return count, nil
}
