package q

import (
	"context"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"io"
	"time"
)

type taskReader struct {
	namespace  *proto.Namespace
	tasks      *storage.TaskStore
	ctx        context.Context
	pipe       chan *proto.Task
	operations []ReadOperation
}

func NewTaskReader(
	ctx context.Context,
	namespace *proto.Namespace,
	tasks *storage.TaskStore,
	ops []ReadOperation,
) *taskReader {
	reader := &taskReader{
		namespace:  namespace,
		tasks:      tasks,
		pipe:       make(chan *proto.Task, 100_000),
		ctx:        ctx,
		operations: ops,
	}
	go reader.startReadLoop()
	return reader
}

func (tr *taskReader) Tasks() TaskStream {
	return tr.pipe
}

func (tr *taskReader) startReadLoop() {
	for {
		select {
		case <-tr.ctx.Done():
			telemetry.Logger.Info("fifo task reader stopped",
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
			time.Sleep(time.Millisecond * 5)
		}
	}
}

// pullAndPush loads ready tasks from storage and delivers them to consumers.
func (tr *taskReader) pullAndPush() (int, error) {
	count := 0
	for _, op := range tr.operations {
		if !op.Ready() {
			continue
		}

		err := op.Run(func(task *proto.Task) error {
			err := tr.push(task)
			if err == nil {
				count++
			}
			return err
		})
		if err != nil {
			return count, err
		}
	}
	return count, nil
}

func (tr *taskReader) push(task *proto.Task) error {
	select {
	case <-tr.ctx.Done():
		return io.EOF

	case tr.pipe <- task:
		return nil
	}
}
