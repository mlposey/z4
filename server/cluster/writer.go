package cluster

import (
	"errors"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/feeds/q"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	pb "google.golang.org/protobuf/proto"
	"sync"
)

type batchWriter struct {
	tw     q.TaskWriter
	ns     *storage.SettingStore
	Handle *LeaderHandle

	acks     []*proto.Ack
	tasks    []*proto.Task
	settings []*proto.QueueConfig

	mu sync.Mutex
	wg sync.WaitGroup
}

func newBatchWriter(
	tw q.TaskWriter,
	ns *storage.SettingStore,
) *batchWriter {
	return &batchWriter{
		tw: tw,
		ns: ns,
	}
}

func (w *batchWriter) Write(logs []*raft.Log) []interface{} {
	w.mu.Lock()
	defer w.mu.Unlock()
	defer w.reset()

	res := make([]interface{}, len(logs))
	err := w.groupByType(logs)
	if err != nil {
		return w.packErrors(err, res)
	}

	var errs [3]error
	w.applySettings(&errs[0])
	w.applyTasks(&errs[1])
	w.applyAcks(&errs[2])

	w.wg.Wait()
	for _, err := range errs {
		if err != nil {
			return w.packErrors(err, res)
		}
	}
	return res
}

func (w *batchWriter) reset() {
	w.acks = nil
	w.settings = nil
	w.tasks = nil
}

func (w *batchWriter) groupByType(logs []*raft.Log) error {
	for _, log := range logs {
		if log.Type != raft.LogCommand {
			continue
		}

		cmd := new(proto.Command)
		err := pb.Unmarshal(log.Data, cmd)
		if err != nil {
			return err
		}

		switch v := cmd.GetCmd().(type) {
		case *proto.Command_Task:
			w.tasks = append(w.tasks, v.Task)

		case *proto.Command_Ack:
			w.acks = append(w.acks, v.Ack)

		case *proto.Command_Queue:
			w.settings = append(w.settings, v.Queue)

		default:
			return errors.New("unknown command type: expected ack or task")
		}
	}
	return nil
}

func (w *batchWriter) packErrors(err error, res []interface{}) []interface{} {
	for i := 0; i < len(res); i++ {
		res[i] = err
	}
	return res
}

func (w *batchWriter) applySettings(err *error) {
	if len(w.settings) == 0 {
		return
	}

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for _, queue := range w.settings {
			*err = w.ns.Save(queue)
			if *err != nil {
				return
			}
		}
	}()
}

func (w *batchWriter) applyTasks(err *error) {
	if len(w.tasks) == 0 {
		return
	}

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		if w.Handle == nil {
			*err = w.tw.Push(w.tasks, true)
		} else {
			*err = w.tw.Push(w.tasks, !w.Handle.IsLeader())
		}
	}()
}

func (w *batchWriter) applyAcks(err *error) {
	if len(w.acks) == 0 {
		return
	}

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		*err = w.tw.Acknowledge(w.acks)
	}()
}
