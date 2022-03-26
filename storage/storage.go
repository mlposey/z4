package storage

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type Task struct {
	ID int64
	TaskDefinition
}

type TaskDefinition struct {
	RunTime  time.Time
	Metadata map[string]string
	Payload  []byte
}

// TaskRange is a query for tasks within a time range.
type TaskRange struct {
	Min time.Time
	Max time.Time
}

// SimpleStore manages a collection of tasks.
//
// This is a very inefficient, in-memory implementation of a task store.
// It should not be used in production.
// It is safe for concurrent use by multiple goroutines.
type SimpleStore struct {
	tasks []Task
	mu    sync.RWMutex
}

func (s *SimpleStore) Save(ctx context.Context, task TaskDefinition) (Task, error) {
	t := Task{
		ID:             rand.Int63(),
		TaskDefinition: task,
	}
	s.mu.Lock()
	s.tasks = append(s.tasks, t)
	s.mu.Unlock()
	return t, nil
}

func (s *SimpleStore) Get(ctx context.Context, tr TaskRange) ([]Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tasks []Task
	for _, task := range s.tasks {
		if task.RunTime.After(tr.Min) && task.RunTime.Before(tr.Max) {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (s *SimpleStore) Delete(ctx context.Context, task Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := -1
	for i, t := range s.tasks {
		if t.ID == task.ID {
			idx = i
			break
		}
	}

	if idx != -1 {
		s.tasks = append(s.tasks[0:idx], s.tasks[idx+1:]...)
	}
	return nil
}
