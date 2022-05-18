package storage

import (
	"bytes"
	"fmt"
	"github.com/cockroachdb/pebble"
	"github.com/dgraph-io/badger/v3"
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
const SequenceLeaseSize = 1000

// NamespaceStore manages persistent storage for namespace configurations.
type NamespaceStore struct {
	Client *BadgerClient
	prefix []byte
}

func NewNamespaceStore(client *BadgerClient) *NamespaceStore {
	return &NamespaceStore{
		Client: client,
		prefix: []byte("namespace#config#"),
	}
}

func (cs *NamespaceStore) Sequence(namespaceID string) (*badger.Sequence, error) {
	key := getSeqKey(namespaceID)
	return cs.Client.DB.GetSequence(key, SequenceLeaseSize)
}

func getSeqKey(namespaceID string) []byte {
	return []byte(fmt.Sprintf("namespace#seq#%s", namespaceID))
}

func (cs *NamespaceStore) Save(namespace *proto.Namespace) error {
	payload, err := pb.Marshal(namespace)
	if err != nil {
		return fmt.Errorf("could not encode namespace: %w", err)
	}
	key := cs.getConfigKey(namespace.GetId())
	return cs.Client.DB2.Set(key, payload, pebble.NoSync)
}

func (cs *NamespaceStore) GetAll() ([]*proto.Namespace, error) {
	telemetry.Logger.Debug("getting all namespace namespaces from DB")

	it := cs.Client.DB2.NewIter(&pebble.IterOptions{})
	defer it.Close()
	it.SeekGE(cs.prefix)

	var namespaces []*proto.Namespace
	for ; it.Valid() && bytes.HasPrefix(it.Key(), cs.prefix); it.Next() {
		namespace := new(proto.Namespace)
		err := pb.Unmarshal(it.Value(), namespace)
		if err != nil {
			telemetry.Logger.Error("failed to load namespace config from database",
				zap.String("key", string(it.Key())),
				zap.Error(err))
		} else {
			namespaces = append(namespaces, namespace)
		}
	}
	return namespaces, nil
}

func (cs *NamespaceStore) Get(id string) (*proto.Namespace, error) {
	telemetry.Logger.Debug("getting namespace config from DB")
	key := cs.getConfigKey(id)
	item, closer, err := cs.Client.DB2.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	namespace := new(proto.Namespace)
	err = pb.Unmarshal(item, namespace)
	return namespace, err
}

func (cs *NamespaceStore) getConfigKey(namespaceID string) []byte {
	return []byte(fmt.Sprintf("namespace#config#%s", namespaceID))
}
