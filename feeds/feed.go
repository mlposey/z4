package feeds

import (
	"context"
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/feeds/q"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"io"
	"time"
)

// Feed provides access to a stream of tasks that are ready to be delivered.
type Feed struct {
	Namespace *storage.SyncedNamespace
	feed      chan *proto.Task
	ctx       context.Context
	ctxCancel context.CancelFunc
	readers   []q.Reader
}

func New(
	namespaceID string,
	db *storage.PebbleClient,
	raft *raft.Raft,
) (*Feed, error) {
	ctx, cancel := context.WithCancel(context.Background())

	f := &Feed{
		Namespace: storage.NewSyncedNamespace(storage.NewNamespaceStore(db), namespaceID, raft),
		feed:      make(chan *proto.Task, 100_000),
		ctx:       ctx,
		ctxCancel: cancel,
	}

	if err := f.Namespace.StartSync(); err != nil {
		return nil, fmt.Errorf("feed creation failed due to namespace error: %w", err)
	}

	tasks := storage.NewTaskStore(db)
	f.readers = q.Readers(tasks, f.Namespace.N)

	go f.startFeed()
	return f, nil
}

func (f *Feed) startFeed() {
	defer close(f.feed)
	telemetry.Logger.Info("feed started",
		zap.String("namespace", f.Namespace.N.GetId()))

	for {
		select {
		case <-f.ctx.Done():
			return
		default:
		}

		pushCount, err := f.pullAndPush()
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
func (f *Feed) pullAndPush() (int, error) {
	count := 0
	for _, reader := range f.readers {
		if !reader.Ready() {
			continue
		}

		err := reader.Read(func(task *proto.Task) error {
			err := f.push(task)
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

func (f *Feed) push(task *proto.Task) error {
	select {
	case <-f.ctx.Done():
		return io.EOF

	case f.feed <- task:
		return nil
	}
}

func (f *Feed) Tasks() q.TaskStream {
	return f.feed
}

// Close stops the feed from listening to ready tasks.
//
// This method must be called once after the feed is no longer
// needed.
func (f *Feed) Close() error {
	f.ctxCancel()
	err := f.Namespace.Close()
	if err != nil {
		return fmt.Errorf("failed to close feed for namespace '%s': %w", f.Namespace.N.GetId(), err)
	}

	telemetry.Logger.Info("feed stopped",
		zap.String("namespace", f.Namespace.N.GetId()))
	return nil
}
