package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	pb "google.golang.org/protobuf/proto"
	"io"
	"time"
)

// BadgerClient manages a connection to a BadgerDB file store.
type BadgerClient struct {
	DB *badger.DB
}

func NewBadgerClient(dataDir string) (*BadgerClient, error) {
	db, err := badger.Open(badger.DefaultOptions(dataDir))
	if err != nil {
		return nil, fmt.Errorf("could not open badger database: %w", err)
	}
	client := &BadgerClient{DB: db}
	go client.handleGC()
	return client, nil
}

func (bc *BadgerClient) handleGC() {
	// TODO: Consider making these values configurable.

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		telemetry.Logger.Debug("running value log garbage collector")

	again:
		err := bc.DB.RunValueLogGC(0.7)
		if err == nil {
			goto again
		}
		telemetry.LastDBGC.SetToCurrentTime()
	}
}

// Close flushes writes to disk.
//
// It must be called when the client is no longer needed or else
// pending writes may be canceled when the application terminates.
func (bc *BadgerClient) Close() error {
	return bc.DB.Close()
}

// NamespaceStore manages persistent storage for namespace configurations.
type NamespaceStore struct {
	Client *BadgerClient
	prefix []byte
}

func NewNamespaceStore(client *BadgerClient) *NamespaceStore {
	return &NamespaceStore{
		Client: client,
		prefix: []byte("namespace#"),
	}
}

func (cs *NamespaceStore) Save(namespace *proto.Namespace) error {
	return cs.Client.DB.Update(func(txn *badger.Txn) error {
		payload, err := json.Marshal(namespace)
		if err != nil {
			return fmt.Errorf("could not encode namespace: %w", err)
		}
		key := cs.getKey(namespace.GetId())
		return txn.Set(key, payload)
	})
}

func (cs *NamespaceStore) GetAll() ([]*proto.Namespace, error) {
	telemetry.Logger.Debug("getting all namespace namespaces from DB")

	var namespaces []*proto.Namespace
	return namespaces, cs.Client.DB.View(func(txn *badger.Txn) error {

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		for it.Seek(cs.prefix); it.ValidForPrefix(cs.prefix); it.Next() {
			item := it.Item()

			_ = item.Value(func(val []byte) error {
				namespace := new(proto.Namespace)
				err := pb.Unmarshal(val, namespace)
				if err != nil {
					telemetry.Logger.Error("failed to load namespace config from database",
						zap.String("key", string(item.Key())))
				} else {
					namespaces = append(namespaces, namespace)
				}
				return err
			})
		}
		return nil
	})
}

func (cs *NamespaceStore) Get(id string) (*proto.Namespace, error) {
	telemetry.Logger.Debug("getting namespace config from DB")
	var namespace *proto.Namespace
	return namespace, cs.Client.DB.View(func(txn *badger.Txn) error {
		key := cs.getKey(id)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			namespace = new(proto.Namespace)
			return pb.Unmarshal(val, namespace)
		})
	})
}

func (cs *NamespaceStore) getKey(namespaceID string) []byte {
	return []byte(fmt.Sprintf("namespace#%s", namespaceID))
}

// TaskStore manages persistent storage for tasks.
type TaskStore struct {
	Client *BadgerClient
}

func NewTaskStore(db *BadgerClient) *TaskStore {
	store := &TaskStore{Client: db}
	return store
}

func (ts *TaskStore) DeleteAll(acks []*proto.Ack) error {
	telemetry.Logger.Debug("deleting task batch from DB", zap.Int("count", len(acks)))
	batch := ts.Client.DB.NewWriteBatch()
	defer batch.Cancel()

	for _, ack := range acks {
		key := getTaskKey(ack.GetNamespace(), ack.GetTaskId())
		err := batch.Delete(key)
		if err != nil {
			return fmt.Errorf("failed to delete task '%s' in batch: %w", ack.GetTaskId(), err)
		}
	}
	return batch.Flush()
}

func (ts *TaskStore) Save(task *proto.Task) error {
	telemetry.Logger.Debug("writing task to DB", zap.Any("task", task))
	return ts.Client.DB.Update(func(txn *badger.Txn) error {
		payload, err := pb.Marshal(task)
		if err != nil {
			return fmt.Errorf("could not encode task: %w", err)
		}
		key := getTaskKey(task.GetNamespace(), task.GetId())
		return txn.Set(key, payload)
	})
}

func (ts *TaskStore) SaveAll(tasks []*proto.Task) error {
	telemetry.Logger.Debug("writing task batch to DB", zap.Int("count", len(tasks)))
	batch := ts.Client.DB.NewWriteBatch()
	defer batch.Cancel()

	for _, task := range tasks {
		payload, err := pb.Marshal(task)
		if err != nil {
			return fmt.Errorf("count not encode task '%s': %w", task.GetId(), err)
		}
		// TODO: Determine if grouping tasks by namespace before writing is beneficial.
		key := getTaskKey(task.GetNamespace(), task.GetId())
		err = batch.Set(key, payload)
		if err != nil {
			return fmt.Errorf("failed to write task '%s' from batch: %w", task.GetId(), err)
		}
	}
	return batch.Flush()
}

func (ts *TaskStore) Get(namespace, id string) (*proto.Task, error) {
	var task *proto.Task
	err := ts.Client.DB.View(func(txn *badger.Txn) error {
		key := getTaskKey(namespace, id)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			task = new(proto.Task)
			return pb.Unmarshal(val, task)
		})
	})
	return task, err
}

func (ts *TaskStore) GetRange(query TaskRange) ([]*proto.Task, error) {
	it, err := ts.IterateRange(query)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	var tasks []*proto.Task
	err = it.ForEach(func(task *proto.Task) error {
		tasks = append(tasks, task)
		return nil
	})
	return tasks, err
}

func (ts *TaskStore) IterateRange(query TaskRange) (*TaskIterator, error) {
	err := query.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks due to invalid query: %w", err)
	}
	return NewTaskIterator(ts.Client, query), nil
}

type TaskIterator struct {
	txn *badger.Txn
	it  *badger.Iterator
	end []byte
}

func NewTaskIterator(client *BadgerClient, query TaskRange) *TaskIterator {
	// TODO: Consider creating another type to pass in instead of BadgerClient.
	// We should try to avoid usage of BadgerClient in other packages as much as possible.

	txn := client.DB.NewTransaction(false)
	it := txn.NewIterator(badger.DefaultIteratorOptions)
	start := getTaskKey(query.Namespace, query.StartID)
	it.Seek(start)

	return &TaskIterator{
		txn: txn,
		it:  it,
		end: getTaskKey(query.Namespace, query.EndID),
	}
}

// ForEach calls handle on the result of Next until an error occurs or no tasks remain.
func (ti *TaskIterator) ForEach(handle func(task *proto.Task) error) error {
	for {
		task, err := ti.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		err = handle(task)
		if err != nil {
			return err
		}
	}
}

func (ti *TaskIterator) Next() (*proto.Task, error) {
	if !ti.it.Valid() {
		return nil, io.EOF
	}

	task, err := ti.peek(true)
	if err != io.EOF {
		ti.it.Next()
	}
	return task, err
}

func (ti *TaskIterator) Peek() (*proto.Task, error) {
	return ti.peek(false)
}

func (ti *TaskIterator) peek(skipCheck bool) (*proto.Task, error) {
	if !skipCheck && !ti.it.Valid() {
		return nil, io.EOF
	}

	item := ti.it.Item()
	if bytes.Compare(item.Key(), ti.end) > 0 {
		return nil, io.EOF
	}

	task := new(proto.Task)
	err := item.Value(func(val []byte) error {
		err := pb.Unmarshal(val, task)
		if err != nil {
			return err
		}

		if task.GetId() == "" {
			return errors.New("invalid task: empty id")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (ti *TaskIterator) Close() error {
	ti.it.Close()
	ti.txn.Discard()
	return nil
}

func getTaskKey(namespaceID, taskID string) []byte {
	return []byte(fmt.Sprintf("task#%s#%s", namespaceID, taskID))
}
