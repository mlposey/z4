package queue

import (
	"context"
	"log"
	"time"
	"z4/storage"
)

type Tasks struct {
	store *storage.SimpleStore
}

func New() *Tasks {
	return &Tasks{store: new(storage.SimpleStore)}
}

func (t *Tasks) Feed(ctx context.Context) <-chan storage.Task {
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
				log.Printf("failed to fetch tasks: %v\n", err)
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
