package storage

import (
	"fmt"
	"github.com/araddon/dateparse"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/auth"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/information_schema"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/telemetry"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"io"
	"time"
)

// TODO: Refine this implementation.
// SQL support is still pretty fragile. We need to increase the robustness
// of the namespace/timestamp indexing and also add functional tests.

const (
	sqlColumnDeliverAt = "deliver_at"
	sqlColumnNamespace = "namespace"
	sqlColumnID        = "id"
	sqlColumnMetadata  = "metadata"
	sqlColumnPayload   = "payload"
)

type WireConfig struct {
	Port  int
	Store *TaskStore
}

func StartWireListener(config WireConfig) error {
	db := NewDatabase("z4")
	const tasksTable = "tasks"
	db.AddTable(NewTable(tasksTable, sql.Schema{
		{Name: sqlColumnNamespace, Type: sql.Text, Nullable: false, Source: tasksTable},
		{Name: sqlColumnID, Type: sql.Text, Nullable: false, Source: tasksTable},
		{Name: sqlColumnDeliverAt, Type: sql.Timestamp, Nullable: false, Source: tasksTable},
		{Name: sqlColumnMetadata, Type: sql.JSON, Nullable: true, Source: tasksTable},
		{Name: sqlColumnPayload, Type: sql.Blob, Nullable: true, Source: tasksTable},
	}, config.Store))

	engine := sqle.NewDefault(
		sql.NewDatabaseProvider(
			db,
			information_schema.NewInformationSchemaDatabase(),
		))

	scfg := server.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("0.0.0.0:%d", config.Port),
		Auth:     &auth.None{},
	}
	s, err := server.NewDefaultServer(scfg, engine)
	if err != nil {
		telemetry.Logger.Fatal("could not start mysql wire server", zap.Error(err))
	}
	err = s.Start()
	if err != nil {
		telemetry.Logger.Fatal("mysql wire server failed", zap.Error(err))
	}
	return nil
}

func RowFromTask(task *proto.Task) sql.Row {
	return sql.NewRow(
		task.GetNamespace(),
		task.GetId(),
		task.GetDeliverAt().AsTime(),
		sql.JSONDocument{Val: task.GetMetadata()},
		task.GetPayload(),
	)
}

type Database struct {
	name   string
	tables map[string]sql.Table
}

func NewDatabase(name string) *Database {
	return &Database{
		name:   name,
		tables: make(map[string]sql.Table),
	}
}

func (d *Database) AddTable(table sql.Table) {
	d.tables[table.Name()] = table
}

func (d *Database) Name() string {
	return d.name
}

func (d *Database) GetTableInsensitive(ctx *sql.Context, tblName string) (sql.Table, bool, error) {
	table, ok := sql.GetTableInsensitive(tblName, d.tables)
	return table, ok, nil
}

func (d *Database) GetTableNames(ctx *sql.Context) ([]string, error) {
	var names []string
	for name := range d.tables {
		names = append(names, name)
	}
	return names, nil
}

type Table struct {
	name    string
	schema  sql.Schema
	filters []sql.Expression
	store   *TaskStore
}

var _ sql.Table = (*Table)(nil)
var _ sql.FilteredTable = (*Table)(nil)

func NewTable(name string, schema sql.Schema, store *TaskStore) *Table {
	return &Table{
		name:   name,
		schema: schema,
		store:  store,
	}
}

func TableWithFilters(table *Table, filters []sql.Expression) *Table {
	return &Table{
		name:    table.name,
		schema:  table.schema,
		filters: filters,
		store:   table.store,
	}
}

func (t *Table) Name() string {
	return t.name
}

func (t *Table) String() string {
	return t.name
}

func (t *Table) Schema() sql.Schema {
	return t.schema
}

func (t *Table) Partitions(context *sql.Context) (sql.PartitionIter, error) {
	// we don't care about partitions. basically no-op'ing these
	return &partitionIterator{}, nil
}

func (t *Table) PartitionRows(context *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	return newRowIterator(t.filters, t.store), nil
}

func (t *Table) HandledFilters(filters []sql.Expression) []sql.Expression {
	return filters
}

func (t *Table) WithFilters(ctx *sql.Context, filters []sql.Expression) sql.Table {
	// I *think* we need to create a copy to avoid race conditions on t.filters
	return TableWithFilters(t, filters)
}

// partitionIterator is a no-op iterator; we don't support partitions
type partitionIterator struct {
	i int
}

func (p *partitionIterator) Close(context *sql.Context) error {
	return nil
}

func (p *partitionIterator) Next() (sql.Partition, error) {
	if p.i == 0 {
		p.i++
		return &Partition{}, nil
	}
	return nil, io.EOF
}

// Partition is a no-op partition; we don't support partitions
type Partition struct {
}

func (p *Partition) Key() []byte {
	return nil
}

type rowIterator struct {
	filters    []sql.Expression
	store      *TaskStore
	rangeStart time.Time
	rangeEnd   time.Time
	namespace  string
	it         *TaskIterator
}

func newRowIterator(filters []sql.Expression, store *TaskStore) *rowIterator {
	it := &rowIterator{
		filters: filters,
		store:   store,
	}
	it.initBounds()
	return it
}

func (r *rowIterator) initBounds() {
	for _, filter := range r.filters {
		sql.Inspect(filter, func(expr sql.Expression) bool {
			if r.detectTimeBounds(expr) {
				return false
			}
			return !r.detectNamespace(expr)
		})
	}

	r.it = NewTaskIterator(r.store.Client, TaskRange{
		Namespace: r.namespace,
		StartID:   NewTaskID(r.rangeStart),
		EndID:     NewTaskID(r.rangeEnd),
	})
}

func (r *rowIterator) detectTimeBounds(f sql.Expression) bool {
	switch v := f.(type) {
	case *expression.Between:
		if r.isFieldExpression(v.Val, sqlColumnDeliverAt) {
			r.rangeStart = r.getTime(v.Lower)
			r.rangeEnd = r.getTime(v.Upper)
			return true
		}

	case *expression.GreaterThan:
		if r.isFieldExpression(v.Left(), sqlColumnDeliverAt) {
			r.rangeStart = r.getTime(v.Right())
			r.rangeEnd = time.Now().Add(time.Hour * 24 * 365 * 10)
			return true
		}

	case *expression.GreaterThanOrEqual:
		if r.isFieldExpression(v.Left(), sqlColumnDeliverAt) {
			r.rangeStart = r.getTime(v.Right())
			r.rangeEnd = time.Now().Add(time.Hour * 24 * 365 * 10)
			return true
		}

	case *expression.LessThan:
		if r.isFieldExpression(v.Left(), sqlColumnDeliverAt) {
			r.rangeStart = ksuid.Nil.Time()
			r.rangeEnd = r.getTime(v.Right())
			return true
		}

	case *expression.LessThanOrEqual:
		if r.isFieldExpression(v.Left(), sqlColumnDeliverAt) {
			r.rangeStart = ksuid.Nil.Time()
			r.rangeEnd = r.getTime(v.Right())
			return true
		}

	case *expression.Equals:
		// TODO: This expression type does not work. No results are returned.
		if r.isFieldExpression(v.Left(), sqlColumnDeliverAt) {
			r.rangeStart = r.getTime(v.Right())
			r.rangeEnd = r.rangeStart
			return true
		}

	default:
		return false
	}
	return false
}

func (r *rowIterator) detectNamespace(f sql.Expression) bool {
	equal, ok := f.(*expression.Equals)
	if !ok {
		return false
	}

	be := equal.BinaryExpression
	if r.isFieldExpression(be.Left, "namespace") {
		r.namespace = r.getNamespace(be.Right)
		return true
	}
	return false
}

func (r *rowIterator) isFieldExpression(f sql.Expression, field string) bool {
	var found bool
	sql.Inspect(f, func(expr sql.Expression) bool {
		if e, ok := expr.(*expression.GetField); ok {
			if e.Name() == field {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func (r *rowIterator) getNamespace(f sql.Expression) string {
	var namespace string

	sql.Inspect(f, func(expr sql.Expression) bool {
		if namespace != "" {
			return false
		}
		lit, ok := expr.(*expression.Literal)
		if !ok {
			return true
		}

		_, ok = lit.Type().(sql.StringType)
		if !ok {
			return true
		}
		namespace = lit.Value().(string)
		return false
	})
	return namespace
}

func (r *rowIterator) getTime(f sql.Expression) time.Time {
	var ts time.Time
	var err error

	sql.Inspect(f, func(expr sql.Expression) bool {
		if !ts.IsZero() {
			return false
		}
		lit, ok := expr.(*expression.Literal)
		if !ok {
			return true
		}

		_, ok = lit.Type().(sql.StringType)
		if !ok {
			return true
		}

		ts, err = dateparse.ParseAny(lit.Value().(string))
		if err != nil {
			fmt.Println("error converting time", err)
			return true
		}
		return false
	})
	return ts
}

func (r *rowIterator) Next() (sql.Row, error) {
	task, err := r.it.Next()
	if err != nil {
		if err != io.EOF {
			telemetry.Logger.Error("encountered error when iterating task range", zap.Error(err))
		}
		return nil, io.EOF
	}
	row := RowFromTask(task)

	for _, filter := range r.filters {
		result, err := filter.Eval(sql.NewEmptyContext(), row)
		if err != nil {
			return nil, err
		}
		result, _ = sql.ConvertToBool(result)
		if result != true {
			return r.Next()
		}
	}

	return row, nil
}

func (r *rowIterator) Close(context *sql.Context) error {
	return r.it.Close()
}
