package q

import (
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"time"
)

type deliveredReader struct {
	namespace *proto.Namespace
	tasks     *storage.TaskStore
	lastRun   time.Time
	interval  time.Duration
	active    bool
	qf        QueryFactory
}

var _ Reader = (*deliveredReader)(nil)

func newDeliveredReader(
	interval time.Duration,
	tasks *storage.TaskStore,
	namespace *proto.Namespace,
	qf QueryFactory,
) *deliveredReader {
	return &deliveredReader{
		namespace: namespace,
		tasks:     tasks,
		lastRun:   time.Time{},
		interval:  interval,
		active:    false,
		qf:        qf,
	}
}

func (d *deliveredReader) Ready() bool {
	return time.Since(d.lastRun) > d.interval && !d.active
}

func (d *deliveredReader) Read(f func(task *proto.Task) error) error {
	d.lastRun = time.Now()
	d.active = true
	go func() {
		err := d.run(f)
		if err != nil {
			telemetry.Logger.Error("redelivery reader failed",
				zap.Error(err),
				zap.String("namespace", d.namespace.GetId()))
		}
	}()
	return nil
}

func (d *deliveredReader) run(f func(task *proto.Task) error) error {
	var tasks []*proto.Task
	err := d.executeQuery(func(task *proto.Task) error {
		if err := f(task); err != nil {
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

	err = d.tasks.SaveAll(tasks, false)
	if err != nil {
		telemetry.Logger.Error("failed to update retry tasks",
			zap.Error(err))
		return err
	}

	d.active = false
	return nil
}

func (d *deliveredReader) executeQuery(handle func(task *proto.Task) error) error {
	ackDeadline := time.Second * time.Duration(d.namespace.GetAckDeadlineSeconds())
	watermark := time.Now().Add(-ackDeadline)
	query := d.qf.Query(d.namespace)
	it := storage.NewTaskIterator(d.tasks.Client, query)

	return it.ForEach(func(task *proto.Task) error {
		if d.retryTask(task, watermark) {
			if err := handle(task); err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *deliveredReader) retryTask(task *proto.Task, watermark time.Time) bool {
	if task.GetLastRetry() == nil {
		return task.GetCreatedAt().AsTime().Before(watermark)
	} else {
		return task.GetLastRetry().AsTime().Before(watermark)
	}
}
