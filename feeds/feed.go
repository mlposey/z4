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
)

// Feed provides access to a stream of tasks that are ready to be delivered.
type Feed struct {
	Namespace      *storage.SyncedNamespace
	feed           chan *proto.Task
	scheduledTasks q.TaskReader
	ctx            context.Context
	ctxCancel      context.CancelFunc
}

func New(
	namespaceID string,
	db *storage.BadgerClient,
	raft *raft.Raft,
) (*Feed, error) {
	ctx, cancel := context.WithCancel(context.Background())

	f := &Feed{
		Namespace: storage.NewSyncedNamespace(storage.NewNamespaceStore(db), namespaceID, raft),
		feed:      make(chan *proto.Task),
		ctx:       ctx,
		ctxCancel: cancel,
	}

	if err := f.Namespace.StartSync(); err != nil {
		return nil, fmt.Errorf("feed creation failed due to namespace error: %w", err)
	}

	tasks := storage.NewTaskStore(db)
	f.scheduledTasks = q.NewScheduledTaskReader(
		ctx,
		f.Namespace.N,
		tasks,
	)

	go f.startFeed()
	return f, nil
}

func (f *Feed) startFeed() {
	defer close(f.feed)
	telemetry.Logger.Info("feed started",
		zap.String("namespace", f.Namespace.N.GetId()))

	scheduled := f.scheduledTasks.Tasks()
	for {
		select {
		case <-f.ctx.Done():
			return

		case task := <-scheduled:
			if err := f.push(task); err != nil {
				return
			}

			// TODO: case task := <-fifoTask
		}
	}
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
