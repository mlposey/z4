package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/telemetry"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"time"
)

type BadgerClient struct {
	db *badger.DB
}

func NewBadgerClient(dataDir string) (*BadgerClient, error) {
	db, err := badger.Open(badger.DefaultOptions(dataDir))
	if err != nil {
		return nil, fmt.Errorf("could not open badger database: %w", err)
	}
	client := &BadgerClient{db: db}
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
		err := bc.db.RunValueLogGC(0.7)
		if err == nil {
			goto again
		}
	}
}

func (bc *BadgerClient) SaveConfig(namespace string, config QueueConfig) error {
	return bc.db.Update(func(txn *badger.Txn) error {
		payload, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("could not encode config: %w", err)
		}
		key := bc.getConfigFQN(namespace)
		return txn.Set(key, payload)
	})
}

func (bc *BadgerClient) GetConfig(namespace string) (QueueConfig, error) {
	telemetry.Logger.Debug("getting config from db")
	var config QueueConfig
	return config, bc.db.View(func(txn *badger.Txn) error {
		key := bc.getConfigFQN(namespace)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &config)
		})
	})
}

func (bc *BadgerClient) Close() error {
	return bc.db.Close()
}

func (bc *BadgerClient) SaveTask(namespace string, task Task) error {
	telemetry.Logger.Debug("writing task to db", zap.Any("task", task))
	return bc.db.Update(func(txn *badger.Txn) error {
		// TODO: Use protobuf for task encode/decode.
		payload, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("could not encode task: %w", err)
		}
		key := bc.getTaskFQN(namespace, task.ID)
		return txn.Set(key, payload)
	})
}

func (bc *BadgerClient) GetTask(query TaskRange) ([]Task, error) {
	var tasks []Task
	err := bc.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		startID := bc.getTaskFQN(query.Namespace, query.MinID())
		endID := bc.getTaskFQN(query.Namespace, query.MaxID())

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

				if !task.ID.IsNil() {
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

func (bc *BadgerClient) getTaskFQN(namespace string, id ksuid.KSUID) []byte {
	return []byte(fmt.Sprintf("%s#task#%s", namespace, id.String()))
}

func (bc *BadgerClient) getConfigFQN(namespace string) []byte {
	return []byte(fmt.Sprintf("%s#config", namespace))
}
