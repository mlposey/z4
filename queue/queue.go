package queue

import (
	"context"
	"go.uber.org/zap"
	"time"
	"z4/storage"
	"z4/telemetry"
)

type Tasks struct {
	store *storage.SimpleStore
}

func New() *Tasks {
	return &Tasks{store: new(storage.SimpleStore)}
}

func (t *Tasks) Feed(ctx context.Context) <-chan storage.Task {
	// TODO: Move goroutine logic out of this method.
	// We have to worry about multiple calls to Feed starting
	// goroutines that duplicate efforts. This logic can be
	// moved to the New method or maybe a Start method. Then,
	// Feed will simply return the task channel.

	delay := time.Millisecond * 10
	var lastRun time.Time
	taskC := make(chan storage.Task)
	go func() {
		for ctx.Err() == nil {
			now := time.Now()
			tasks, err := t.store.Get(ctx, storage.TaskRange{
				Min: lastRun,
				Max: now,
			})
			lastRun = now

			if err != nil {
				telemetry.Logger.Error("failed to fetch tasks",
					zap.Error(err))
			}

			for _, task := range tasks {
				taskC <- task
			}

			sleep := time.Now().Sub(lastRun)
			if sleep < delay {
				time.Sleep(delay - sleep)
			}
		}
		close(taskC)
	}()
	return taskC
}

func (t *Tasks) Add(ctx context.Context, def storage.TaskDefinition) (storage.Task, error) {
	return t.store.Save(ctx, def)
}

func (t *Tasks) Remove(ctx context.Context, task storage.Task) error {
	return t.store.Delete(ctx, task)
}
