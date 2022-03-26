package storage

import (
	"encoding/json"
	"fmt"
	badger "github.com/dgraph-io/badger/v3"
	"github.com/mlposey/z4/telemetry"
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

func (bc *BadgerClient) Close() error {
	return bc.db.Close()
}

func (bc *BadgerClient) Save(namespace string, task Task) error {
	return bc.db.Update(func(txn *badger.Txn) error {
		// TODO: Use protobuf for task encode/decode.
		payload, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("could not encode task: %w", err)
		}
		key := []byte(bc.getKeyFQN(namespace, task.ID))
		return txn.Set(key, payload)
	})
}

func (bc *BadgerClient) Get(query TaskRange) ([]Task, error) {
	var tasks []Task
	err := bc.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		startID := bc.getKeyFQN(query.Namespace, query.MinID())
		endID := bc.getKeyFQN(query.Namespace, query.MaxID())

		it.Seek([]byte(startID))
		for it.Valid() {
			item := it.Item()
			if string(item.Key()) > endID {
				return nil
			}

			err := item.Value(func(val []byte) error {
				var task Task
				err := json.Unmarshal(val, &task)
				if err != nil {
					return err
				}

				tasks = append(tasks, task)
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

func (bc *BadgerClient) getKeyFQN(namespace string, id string) string {
	return fmt.Sprintf("%s#%s", namespace, id)
}
