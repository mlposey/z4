package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	pb "google.golang.org/protobuf/proto"
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

// ConfigStore manages persistent storage for feed configurations.
type ConfigStore struct {
	Client *BadgerClient
}

func (cs *ConfigStore) Save(namespace string, config FeedConfig) error {
	return cs.Client.DB.Update(func(txn *badger.Txn) error {
		payload, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("could not encode config: %w", err)
		}
		key := cs.getConfigFQN(namespace)
		return txn.Set(key, payload)
	})
}

func (cs *ConfigStore) Get(namespace string) (FeedConfig, error) {
	telemetry.Logger.Debug("getting config from DB")
	var config FeedConfig
	return config, cs.Client.DB.View(func(txn *badger.Txn) error {
		key := cs.getConfigFQN(namespace)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &config)
		})
	})
}

func (cs *ConfigStore) getConfigFQN(namespace string) []byte {
	return []byte(fmt.Sprintf("%s#config", namespace))
}

// TaskStore manages persistent storage for tasks.
type TaskStore struct {
	Client *BadgerClient
}

func NewTaskStore(db *BadgerClient) *TaskStore {
	store := &TaskStore{Client: db}
	return store
}

func (ts *TaskStore) Save(task *proto.Task) error {
	telemetry.Logger.Debug("writing task to DB", zap.Any("task", task))
	return ts.Client.DB.Update(func(txn *badger.Txn) error {
		payload, err := pb.Marshal(task)
		if err != nil {
			return fmt.Errorf("could not encode task: %w", err)
		}
		key := ts.getTaskFQN(task.GetNamespace(), task.GetId())
		return txn.Set(key, payload)
	})
}

func (ts *TaskStore) SaveAll(tasks []*proto.Task) error {
	telemetry.Logger.Debug("writing task batch to DB", zap.Any("count", len(tasks)))
	batch := ts.Client.DB.NewWriteBatch()
	defer batch.Cancel()

	for _, task := range tasks {
		payload, err := pb.Marshal(task)
		if err != nil {
			return fmt.Errorf("count not encode task '%s': %w", task.GetId(), err)
		}
		// TODO: Determine if grouping tasks by namespace before writing is beneficial.
		key := ts.getTaskFQN(task.GetNamespace(), task.GetId())
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
		key := ts.getTaskFQN(namespace, id)
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
	err := query.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks due to invalid query: %w", err)
	}
	return ts.get(query)
}

func (ts *TaskStore) get(query TaskRange) ([]*proto.Task, error) {
	var tasks []*proto.Task
	err := ts.Client.DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		startID := ts.getTaskFQN(query.Namespace, query.StartID)
		endID := ts.getTaskFQN(query.Namespace, query.EndID)

		it.Seek(startID)
		for ; it.Valid(); it.Next() {
			item := it.Item()
			if bytes.Compare(item.Key(), endID) > 0 {
				return nil
			}

			err := item.Value(func(val []byte) error {
				task := new(proto.Task)
				err := pb.Unmarshal(val, task)
				if err != nil {
					return err
				}

				if task.GetId() != "" {
					tasks = append(tasks, task)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return tasks, err
}

func (ts *TaskStore) getTaskFQN(namespace string, id string) []byte {
	return []byte(fmt.Sprintf("%s#task#%s", namespace, id))
}
