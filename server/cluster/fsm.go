package cluster

import (
	"encoding/binary"
	"errors"
	"github.com/cockroachdb/pebble"
	"github.com/hashicorp/raft"
	"github.com/mlposey/z4/feeds/q"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	pb "google.golang.org/protobuf/proto"
	"io"
	"sync"
)

// stateMachine uses raft logs to modify the task database.
type stateMachine struct {
	db     *pebble.DB
	writer q.TaskWriter
	ns     *storage.NamespaceStore
	handle *LeaderHandle
}

func newFSM(
	db *pebble.DB,
	ts q.TaskWriter,
	ns *storage.NamespaceStore,
) *stateMachine {
	// TODO: Do not pass a *badger.DB directly. Create a new type.
	return &stateMachine{
		db:     db,
		writer: ts,
		ns:     ns,
	}
}

func (f *stateMachine) SetHandle(handle *LeaderHandle) {
	f.handle = handle
}

func (f *stateMachine) ApplyBatch(logs []*raft.Log) []interface{} {
	telemetry.ReceivedLogs.Add(float64(len(logs)))
	res := make([]interface{}, len(logs))

	batch, err := f.splitLogs(logs)
	if err != nil {
		return f.packErrors(err, res)
	}

	var wg sync.WaitGroup
	var errs [4]error

	f.applyNamespaces(batch.Namespaces, &wg, &errs[0])
	f.applyTasks(batch.Tasks, &wg, &errs[1])
	f.applyAcks(batch.Acks, &wg, &errs[2])
	f.applyPurges(batch.Purges, &wg, &errs[3])

	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return f.packErrors(err, res)
		}
	}

	telemetry.AppliedLogs.Add(float64(batch.Size()))
	return res
}

func (f *stateMachine) splitLogs(logs []*raft.Log) (*splitBatch, error) {
	sb := new(splitBatch)
	for _, log := range logs {
		cmd := new(proto.Command)
		err := pb.Unmarshal(log.Data, cmd)
		if err != nil {
			return nil, err
		}

		switch v := cmd.GetCmd().(type) {
		case *proto.Command_Task:
			sb.Tasks = append(sb.Tasks, v.Task)

		case *proto.Command_Ack:
			sb.Acks = append(sb.Acks, v.Ack)

		case *proto.Command_Namespace:
			sb.Namespaces = append(sb.Namespaces, v.Namespace)

		case *proto.Command_Purge:
			sb.Purges = append(sb.Purges, v.Purge)

		default:
			return nil, errors.New("unknown command type: expected ack or task")
		}
	}
	return sb, nil
}

type splitBatch struct {
	Acks       []*proto.Ack
	Tasks      []*proto.Task
	Purges     []*proto.PurgeTasksRequest
	Namespaces []*proto.Namespace
}

func (s *splitBatch) Size() int {
	return len(s.Acks) + len(s.Tasks) + len(s.Purges) + len(s.Namespaces)
}

func (f *stateMachine) applyNamespaces(namespaces []*proto.Namespace, wg *sync.WaitGroup, err *error) {
	if len(namespaces) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// TODO: If multiple versions of same namespace, pick most recent.
			// Namespace updates should happen infrequently enough that
			// saving them individually rather than using a batch should
			// be more performant.
			for _, namespace := range namespaces {
				*err = f.ns.Save(namespace)
				if *err != nil {
					return
				}
			}
		}()
	}
}

func (f *stateMachine) applyTasks(tasks []*proto.Task, wg *sync.WaitGroup, err *error) {
	if len(tasks) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if f.handle == nil {
				*err = f.writer.Push(tasks, true)
			} else {
				*err = f.writer.Push(tasks, !f.handle.IsLeader())
			}
		}()
	}
}

func (f *stateMachine) applyAcks(acks []*proto.Ack, wg *sync.WaitGroup, err *error) {
	if len(acks) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			*err = f.writer.Acknowledge(acks)
		}()
	}
}

func (f *stateMachine) applyPurges(purges []*proto.PurgeTasksRequest, wg *sync.WaitGroup, err *error) {
	if len(purges) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, req := range purges {
				*err = f.writer.PurgeTasks(req.GetNamespaceId())
				if *err != nil {
					return
				}
			}
		}()
	}
}

func (f *stateMachine) packErrors(err error, res []interface{}) []interface{} {
	for i := 0; i < len(res); i++ {
		res[i] = err
	}
	return res
}

func (f *stateMachine) Apply(log *raft.Log) interface{} {
	// Internally, this method should never be invoked. The raft package should
	// use our ApplyBatch method instead in order to speed up write performance.
	return errors.New("unexpected call to Apply")
}

func (f *stateMachine) Snapshot() (raft.FSMSnapshot, error) {
	telemetry.Logger.Info("taking fsm snapshot")
	return &snapshot{db: f.db.NewSnapshot()}, nil
}

func (f *stateMachine) Restore(snapshot io.ReadCloser) error {
	telemetry.Logger.Info("restoring fsm from snapshot")
	// TODO: Delete the existing database before restoring.

	batch := f.db.NewBatch()
	l := make([]byte, 4)
	for {
		_, err := snapshot.Read(l)
		if err != nil {
			if err == io.EOF {
				break
			}
			_ = batch.Close()
			return err
		}

		key := make([]byte, binary.BigEndian.Uint32(l))
		_, err = snapshot.Read(key)
		if err != nil {
			if err == io.EOF {
				break
			}
			_ = batch.Close()
			return err
		}

		_, err = snapshot.Read(l)
		if err != nil {
			if err == io.EOF {
				break
			}
			_ = batch.Close()
			return err
		}

		value := make([]byte, binary.BigEndian.Uint32(l))
		_, err = snapshot.Read(value)
		if err != nil {
			if err == io.EOF {
				break
			}
			_ = batch.Close()
			return err
		}

		err = batch.Set(key, value, pebble.Sync)
		if err != nil {
			_ = batch.Close()
			return err
		}
	}
	return batch.Commit(pebble.Sync)
}

type snapshot struct {
	db *pebble.Snapshot
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	telemetry.Logger.Info("persisting fsm snapshot")

	it := s.db.NewIter(&pebble.IterOptions{})
	l := make([]byte, 4)
	for it.Valid() {
		binary.BigEndian.PutUint32(l, uint32(len(it.Key())))
		_, err := sink.Write(l)
		if err != nil {
			return err
		}

		_, err = sink.Write(it.Key())
		if err != nil {
			return err
		}

		binary.BigEndian.PutUint32(l, uint32(len(it.Value())))
		_, err = sink.Write(l)
		if err != nil {
			return err
		}

		_, err = sink.Write(it.Value())
		if err != nil {
			return err
		}
	}
	telemetry.LastFSMSnapshot.SetToCurrentTime()
	return it.Close()
}

func (s *snapshot) Release() {
	_ = s.db.Close()
}
