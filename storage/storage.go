package storage

import (
	"context"
	"math/rand"
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

type TaskRange struct {
	Min time.Time
	Max time.Time
}

type SimpleStore struct {
	tasks []Task
}

func (s *SimpleStore) Save(ctx context.Context, task TaskDefinition) (Task, error) {
	t := Task{
		ID:             rand.Int63(),
		TaskDefinition: task,
	}
	s.tasks = append(s.tasks, t)
	return t, nil
}

func (s *SimpleStore) Get(ctx context.Context, tr TaskRange) ([]Task, error) {
	var tasks []Task
	for _, task := range s.tasks {
		if task.RunTime.After(tr.Min) && task.RunTime.Before(tr.Max) {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (s *SimpleStore) Delete(ctx context.Context, task Task) error {
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
