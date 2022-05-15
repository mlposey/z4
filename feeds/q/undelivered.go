package q

import (
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"io"
)

type undeliveredReader struct {
	namespace *proto.Namespace
	tasks     *storage.TaskStore
	qf        PreemptiveQueryFactory
	check     Checkpointer
}

var _ ReadOperation = (*undeliveredReader)(nil)

func newUndeliveredReader(
	tasks *storage.TaskStore,
	namespace *proto.Namespace,
	qf PreemptiveQueryFactory,
	check Checkpointer,
) *undeliveredReader {
	return &undeliveredReader{
		namespace: namespace,
		tasks:     tasks,
		qf:        qf,
		check:     check,
	}
}

func (u *undeliveredReader) Ready() bool {
	return true
}

func (u *undeliveredReader) Run(f func(task *proto.Task) error) error {
	var count int
	err := u.executeQuery(func(task *proto.Task) error {
		if err := f(task); err != nil {
			return err
		}
		u.check.Set(u.namespace, task)
		count++
		return nil
	})
	u.qf.Inform(count)

	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func (u *undeliveredReader) executeQuery(handle func(task *proto.Task) error) error {
	query := u.qf.Query(u.namespace)
	it := storage.NewTaskIterator(u.tasks.Client, query)
	return it.ForEach(handle)
}
