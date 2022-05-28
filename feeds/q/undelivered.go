package q

import (
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"io"
)

type undeliveredReader struct {
	settings *proto.QueueConfig
	tasks    *storage.TaskStore
	qf       QueryFactory
	check    Checkpointer
}

var _ Reader = (*undeliveredReader)(nil)

func newUndeliveredReader(
	tasks *storage.TaskStore,
	settings *proto.QueueConfig,
	qf QueryFactory,
	check Checkpointer,
) *undeliveredReader {
	return &undeliveredReader{
		settings: settings,
		tasks:    tasks,
		qf:       qf,
		check:    check,
	}
}

func (u *undeliveredReader) Ready() bool {
	return true
}

func (u *undeliveredReader) Read(f func(task *proto.Task) error) error {
	var count int
	err := u.executeQuery(func(task *proto.Task) error {
		if err := f(task); err != nil {
			return err
		}
		u.check.Set(u.settings, task)
		count++
		return nil
	})

	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func (u *undeliveredReader) executeQuery(handle func(task *proto.Task) error) error {
	query := u.qf.Query(u.settings)
	it := storage.NewTaskIterator(u.tasks.Client, query)
	defer it.Close()
	return it.ForEach(handle)
}
