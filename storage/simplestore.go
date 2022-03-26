package storage

import (
	"context"
	"sync"
)

// SimpleStore manages a collection of tasks.
//
// This is a very inefficient, in-memory implementation of a storage store.
// It should not be used in production.
// It is safe for concurrent use by multiple goroutines.
type SimpleStore struct {
	tasks []Task
	mu    sync.RWMutex
}

func (s *SimpleStore) Save(ctx context.Context, task TaskDefinition) (Task, error) {
	t := NewTask(task)
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
