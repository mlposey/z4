package storage

import (
	"context"
	"go.uber.org/zap"
	"time"
	"z4/telemetry"
)

type Queue struct {
	store *SimpleStore
}

func New() *Queue {
	return &Queue{store: new(SimpleStore)}
}

func (t *Queue) Feed(ctx context.Context) <-chan Task {
	// TODO: Move goroutine logic out of this method.
	// We have to worry about multiple calls to Feed starting
	// goroutines that duplicate efforts. This logic can be
	// moved to the New method or maybe a Start method. Then,
	// Feed will simply return the task channel.

	delay := time.Millisecond * 10
	var lastRun time.Time
	taskC := make(chan Task)
	go func() {
		for ctx.Err() == nil {
			now := time.Now()
			tasks, err := t.store.Get(ctx, TaskRange{
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

func (t *Queue) Add(ctx context.Context, def TaskDefinition) (Task, error) {
	return t.store.Save(ctx, def)
}

func (t *Queue) Remove(ctx context.Context, task Task) error {
	return t.store.Delete(ctx, task)
}
