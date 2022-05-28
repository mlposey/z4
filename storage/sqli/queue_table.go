package sqli

import (
	"fmt"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"io"
)

const (
	queueTableID  = "queue"
	queueColumnID = "id"

	queueColumnLastScheduled = "last_delivered_scheduled_task"
	queueColumnLastQueued    = "last_delivered_queued_task"

	queueColumnAckDeadline = "ack_deadline_seconds"
)

type queueTable struct {
	name   string
	schema sql.Schema
	store  *storage.SettingStore
}

var _ sql.Table = (*queueTable)(nil)

func newQueueTable(store *storage.SettingStore) *queueTable {
	return &queueTable{
		name: queueTableID,
		schema: sql.Schema{
			{Name: queueColumnID, Type: sql.Text, Nullable: false, Source: queueTableID},
			{Name: queueColumnLastScheduled, Type: sql.Text, Nullable: true, Source: queueTableID},
			{Name: queueColumnLastQueued, Type: sql.Text, Nullable: true, Source: queueTableID},
			{Name: queueColumnAckDeadline, Type: sql.Uint32, Nullable: false, Source: queueTableID},
		},
		store: store,
	}
}

func (t *queueTable) Name() string {
	return t.name
}

func (t *queueTable) String() string {
	return t.name
}

func (t *queueTable) Schema() sql.Schema {
	return t.schema
}

func (t *queueTable) Partitions(context *sql.Context) (sql.PartitionIter, error) {
	// we don't care about partitions. basically no-op'ing these
	return &noOpPartitionIterator{}, nil
}

func (t *queueTable) PartitionRows(context *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	return newQueueTableIterator(t.store)
}

type queueTableIterator struct {
	items []*proto.QueueConfig
	i     int
}

func newQueueTableIterator(store *storage.SettingStore) (*queueTableIterator, error) {
	items, err := store.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load queue settings from database: %w", err)
	}
	return &queueTableIterator{
		items: items,
		i:     0,
	}, nil
}

func (r *queueTableIterator) Next() (sql.Row, error) {
	if r.i >= len(r.items) {
		return nil, io.EOF
	}
	row := r.items[r.i]
	r.i++
	return r.rowFromSettings(row), nil
}

// rowFromTask converts a task into a sql row.
func (r *queueTableIterator) rowFromSettings(settings *proto.QueueConfig) sql.Row {
	return sql.NewRow(
		settings.GetId(),
		settings.GetLastDeliveredScheduledTask(),
		settings.GetLastDeliveredQueuedTask(),
		settings.GetAckDeadlineSeconds(),
	)
}

func (r *queueTableIterator) Close(context *sql.Context) error {
	return nil
}
