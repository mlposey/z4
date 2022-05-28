package storage

import (
	"bytes"
	"fmt"
	"github.com/cockroachdb/pebble"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	pb "google.golang.org/protobuf/proto"
)

// SequenceLeaseSize is the number of indexes leased at a time.
//
// If the application does not gracefully shut down, this number
// represents the maximum number of indexes that will be "lost".
// Lost indexes are just numbers we will not be able to assign
// to new tasks; they don't represent lost data.
//
// A higher number will permit faster index assignment but at the
// cost of more lost indexes during application crashes.
const SequenceLeaseSize = 10_000

// SettingStore manages persistent storage for queue settings.
type SettingStore struct {
	Client *PebbleClient
	prefix []byte
}

func NewSettingStore(client *PebbleClient) *SettingStore {
	return &SettingStore{
		Client: client,
		prefix: []byte("queue#config#"),
	}
}

func (s *SettingStore) Sequence(queue string) (*Sequence, error) {
	key := getSeqKey(queue)
	return NewSequence(s.Client.DB, key, SequenceLeaseSize)
}

func getSeqKey(queue string) []byte {
	return []byte(fmt.Sprintf("queue#seq#%s", queue))
}

func (s *SettingStore) Save(settings *proto.QueueConfig) error {
	payload, err := pb.Marshal(settings)
	if err != nil {
		return fmt.Errorf("could not encode queue settings: %w", err)
	}
	key := s.getConfigKey(settings.GetId())
	return s.Client.DB.Set(key, payload, pebble.NoSync)
}

func (s *SettingStore) GetAll() ([]*proto.QueueConfig, error) {
	telemetry.Logger.Debug("getting all queue settings from DB")

	it := s.Client.DB.NewIter(&pebble.IterOptions{})
	defer it.Close()
	it.SeekGE(s.prefix)

	var queues []*proto.QueueConfig
	for ; it.Valid() && bytes.HasPrefix(it.Key(), s.prefix); it.Next() {
		settings := new(proto.QueueConfig)
		err := pb.Unmarshal(it.Value(), settings)
		if err != nil {
			telemetry.Logger.Error("failed to load queue settings from database",
				zap.String("key", string(it.Key())),
				zap.Error(err))
		} else {
			queues = append(queues, settings)
		}
	}
	return queues, nil
}

func (s *SettingStore) Get(queue string) (*proto.QueueConfig, error) {
	telemetry.Logger.Debug("getting queue settings from DB")
	key := s.getConfigKey(queue)
	item, closer, err := s.Client.DB.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	settings := new(proto.QueueConfig)
	err = pb.Unmarshal(item, settings)
	return settings, err
}

func (s *SettingStore) getConfigKey(queue string) []byte {
	return []byte(fmt.Sprintf("queue#config#%s", queue))
}
