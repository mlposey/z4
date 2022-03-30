package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"time"
)

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
	}
}

func (bc *BadgerClient) Close() error {
	return bc.DB.Close()
}

type ConfigStore struct {
	Client *BadgerClient
}

func (cs *ConfigStore) Save(namespace string, config QueueConfig) error {
	return cs.Client.DB.Update(func(txn *badger.Txn) error {
		payload, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("could not encode config: %w", err)
		}
		key := cs.getConfigFQN(namespace)
		return txn.Set(key, payload)
	})
}

func (cs *ConfigStore) Get(namespace string) (QueueConfig, error) {
	telemetry.Logger.Debug("getting config from DB")
	var config QueueConfig
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

type TaskStore struct {
	Client *BadgerClient
}

func (ts *TaskStore) Save(task Task) error {
	telemetry.Logger.Debug("writing task to DB", zap.Any("task", task))
	return ts.Client.DB.Update(func(txn *badger.Txn) error {
		// TODO: Use protobuf for task encode/decode.
		payload, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("could not encode task: %w", err)
		}
		key := ts.getTaskFQN(task.Namespace, task.ID)
		return txn.Set(key, payload)
	})
}

func (ts *TaskStore) Get(query TaskRange) ([]Task, error) {
	var tasks []Task
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
				var task Task
				err := json.Unmarshal(val, &task)
				if err != nil {
					return err
				}

				if task.ID != "" {
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
