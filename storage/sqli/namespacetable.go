package sqli

import (
	"fmt"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"io"
)

const (
	namespaceTableID           = "namespace"
	namespaceColumnID          = "id"
	namespaceColumnLastTask    = "last_scheduled_task_id"
	namespaceColumnLastIndex   = "last_queued_task_index"
	namespaceColumnAckDeadline = "ack_deadline_seconds"
)

type namespaceTable struct {
	name       string
	schema     sql.Schema
	namespaces *storage.NamespaceStore
}

var _ sql.Table = (*namespaceTable)(nil)

func newNamespaceTable(namespaces *storage.NamespaceStore) *namespaceTable {
	return &namespaceTable{
		name: namespaceTableID,
		schema: sql.Schema{
			{Name: namespaceColumnID, Type: sql.Text, Nullable: false, Source: namespaceTableID},
			{Name: namespaceColumnLastTask, Type: sql.Text, Nullable: true, Source: namespaceTableID},
			{Name: namespaceColumnLastIndex, Type: sql.Uint64, Nullable: true, Source: namespaceTableID},
			{Name: namespaceColumnAckDeadline, Type: sql.Uint32, Nullable: false, Source: namespaceTableID},
		},
		namespaces: namespaces,
	}
}

func (t *namespaceTable) Name() string {
	return t.name
}

func (t *namespaceTable) String() string {
	return t.name
}

func (t *namespaceTable) Schema() sql.Schema {
	return t.schema
}

func (t *namespaceTable) Partitions(context *sql.Context) (sql.PartitionIter, error) {
	// we don't care about partitions. basically no-op'ing these
	return &noOpPartitionIterator{}, nil
}

func (t *namespaceTable) PartitionRows(context *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	return newNamespaceTableIterator(t.namespaces)
}

type namespaceTableIterator struct {
	items []*proto.Namespace
	i     int
}

func newNamespaceTableIterator(namespaces *storage.NamespaceStore) (*namespaceTableIterator, error) {
	items, err := namespaces.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load namespaces from database: %w", err)
	}
	return &namespaceTableIterator{
		items: items,
		i:     0,
	}, nil
}

func (r *namespaceTableIterator) Next() (sql.Row, error) {
	if r.i >= len(r.items) {
		return nil, io.EOF
	}
	row := r.items[r.i]
	r.i++
	return r.rowFromNamespace(row), nil
}

// rowFromTask converts a task into a sql row.
func (r *namespaceTableIterator) rowFromNamespace(namespace *proto.Namespace) sql.Row {
	return sql.NewRow(
		namespace.GetId(),
		namespace.GetLastTask(),
		namespace.GetLastIndex(),
		namespace.GetAckDeadlineSeconds(),
	)
}

func (r *namespaceTableIterator) Close(context *sql.Context) error {
	return nil
}
