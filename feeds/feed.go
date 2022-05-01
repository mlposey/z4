package feeds

import (
	"fmt"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"time"
)

type TaskStream <-chan *proto.Task

// Feed provides access to a stream of tasks that are ready to be delivered.
type Feed struct {
	tasks       *storage.TaskStore
	namespace   *storage.SyncedNamespace
	feed        chan *proto.Task
	close       chan bool
	ackDeadline time.Duration
}

func New(
	namespaceID string,
	db *storage.BadgerClient,
	ackDeadline time.Duration,
) (*Feed, error) {
	q := &Feed{
		tasks:       storage.NewTaskStore(db),
		namespace:   storage.NewSyncedNamespace(storage.NewNamespaceStore(db), namespaceID),
		feed:        make(chan *proto.Task),
		close:       make(chan bool),
		ackDeadline: ackDeadline,
	}

	if err := q.namespace.StartSync(); err != nil {
		return nil, fmt.Errorf("feed creation failed due to namespace error: %w", err)
	}

	go q.startFeed()
	return q, nil
}

func (f *Feed) startFeed() {
	defer close(f.feed)
	telemetry.Logger.Info("feed started",
		zap.String("namespace", f.namespace.N.GetId()))

	for {
		select {
		case <-f.close:
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
			time.Sleep(time.Millisecond * 50)
			continue
		}
	}
}

// pullAndPush loads ready tasks from storage and delivers them to consumers.
func (f *Feed) pullAndPush() (int, error) {
	lastDeliveredTaskID := f.namespace.N.LastDeliveredTask
	dc, err1 := f.processDelivered(lastDeliveredTaskID)
	uc, err2 := f.processUndelivered(lastDeliveredTaskID)
	return dc + uc, multierr.Combine(err1, err2)
}

func (f *Feed) push(task *proto.Task) error {
	select {
	case <-f.close:
		return io.EOF

	case f.feed <- task:
		return nil
	}
}

func (f *Feed) processDelivered(lastID string) (int, error) {
	df := &deliveredTaskFetcher{
		Tasks:           f.tasks,
		LastDeliveredID: lastID,
		Namespace:       f.namespace.N.GetId(),
		AckDeadline:     f.ackDeadline,
	}

	var tasks []*proto.Task
	err := df.Process(func(task *proto.Task) error {
		if err := f.push(task); err != nil {
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
	err = f.tasks.SaveAll(tasks)
	if err != nil {
		telemetry.Logger.Error("failed to update retry tasks",
			zap.Error(err))
		return len(tasks), err
	}
	return len(tasks), nil
}

func (f *Feed) processUndelivered(lastID string) (int, error) {
	uf := &undeliveredTaskFetcher{
		Tasks:     f.tasks,
		StartID:   lastID,
		Namespace: f.namespace.N.GetId(),
	}

	var count int
	err := uf.Process(func(task *proto.Task) error {
		if err := f.push(task); err != nil {
			return err
		}
		f.namespace.N.LastDeliveredTask = task.GetId()
		count++
		return nil
	})

	if err != nil && err != io.EOF {
		return count, err
	}
	return count, nil
}

func (f *Feed) Tasks() TaskStream {
	return f.feed
}

func (f *Feed) Namespace() string {
	return f.namespace.N.GetId()
}

// Close stops the feed from listening to ready tasks.
//
// This method must be called once after the feed is no longer
// needed.
func (f *Feed) Close() error {
	f.close <- true
	err := f.namespace.Close()
	if err != nil {
		return fmt.Errorf("failed to close feed for namespace '%s': %w", f.namespace, err)
	}

	telemetry.Logger.Info("feed stopped",
		zap.String("namespace", f.namespace.N.GetId()))
	return nil
}
