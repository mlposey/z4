package sqli

import (
	"fmt"
	"github.com/araddon/dateparse"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/mlposey/z4/iden"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"io"
	"time"
)

const (
	taskTableID         = "task"
	taskColumnDeliverAt = "deliver_at"
	taskColumnLastRetry = "last_retry"
	taskColumnNamespace = "namespace"
	taskColumnID        = "id"
	taskColumnMetadata  = "metadata"
	taskColumnPayload   = "payload"
)

type taskTable struct {
	name    string
	schema  sql.Schema
	filters []sql.Expression
	store   *storage.TaskStore
}

var _ sql.Table = (*taskTable)(nil)
var _ sql.FilteredTable = (*taskTable)(nil)

func newTaskTable(tasks *storage.TaskStore) *taskTable {
	return &taskTable{
		name: taskTableID,
		schema: sql.Schema{
			{Name: taskColumnNamespace, Type: sql.Text, Nullable: false, Source: taskTableID},
			{Name: taskColumnID, Type: sql.Text, Nullable: false, Source: taskTableID},
			{Name: taskColumnDeliverAt, Type: sql.Timestamp, Nullable: false, Source: taskTableID},
			{Name: taskColumnLastRetry, Type: sql.Timestamp, Nullable: true, Source: taskTableID},
			{Name: taskColumnMetadata, Type: sql.JSON, Nullable: true, Source: taskTableID},
			{Name: taskColumnPayload, Type: sql.Blob, Nullable: true, Source: taskTableID},
		},
		store: tasks,
	}
}

func newTaskTableWithFilters(table *taskTable, filters []sql.Expression) *taskTable {
	return &taskTable{
		name:    table.name,
		schema:  table.schema,
		filters: filters,
		store:   table.store,
	}
}

func (t *taskTable) Name() string {
	return t.name
}

func (t *taskTable) String() string {
	return t.name
}

func (t *taskTable) Schema() sql.Schema {
	return t.schema
}

func (t *taskTable) Partitions(context *sql.Context) (sql.PartitionIter, error) {
	// we don't care about partitions. basically no-op'ing these
	return &noOpPartitionIterator{}, nil
}

func (t *taskTable) PartitionRows(context *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	return newTaskTableIterator(t.filters, t.store)
}

func (t *taskTable) HandledFilters(filters []sql.Expression) []sql.Expression {
	return filters
}

func (t *taskTable) WithFilters(ctx *sql.Context, filters []sql.Expression) sql.Table {
	// I *think* we need to create a copy to avoid race conditions on t.filters
	return newTaskTableWithFilters(t, filters)
}

type taskTableIterator struct {
	filters        []sql.Expression
	store          *storage.TaskStore
	rangeStart     time.Time
	rangeEnd       time.Time
	namespace      string
	namespaceFound bool
	it             *storage.TaskIterator
}

func newTaskTableIterator(filters []sql.Expression, store *storage.TaskStore) (*taskTableIterator, error) {
	it := &taskTableIterator{
		filters: filters,
		store:   store,
	}
	err := it.initBounds()
	if err != nil {
		return nil, err
	}
	return it, nil
}

func (r *taskTableIterator) initBounds() error {
	for _, filter := range r.filters {
		sql.Inspect(filter, func(expr sql.Expression) bool {
			if r.detectTimeBounds(expr) {
				return false
			}
			return !r.detectNamespace(expr)
		})
	}

	if r.rangeStart.IsZero() || r.rangeEnd.IsZero() {
		return fmt.Errorf("invalid or missing range query for field: %s", taskColumnDeliverAt)
	}

	if !r.namespaceFound {
		return fmt.Errorf("missing required namespace in query")
	}

	r.it = storage.NewTaskIterator(r.store.Client, &storage.ScheduledRange{
		Namespace: r.namespace,
		StartID:   iden.New(r.rangeStart, 0),
		EndID:     iden.New(r.rangeEnd, 0),
	})
	return nil
}

func (r *taskTableIterator) detectTimeBounds(f sql.Expression) bool {
	switch v := f.(type) {
	case *expression.Between:
		if r.isFieldExpression(v.Val, taskColumnDeliverAt) {
			r.rangeStart = r.getTime(v.Lower)
			r.rangeEnd = r.getTime(v.Upper)
			return true
		}

	case *expression.GreaterThan:
		if r.isFieldExpression(v.Left(), taskColumnDeliverAt) {
			r.rangeStart = r.getTime(v.Right())
			r.rangeEnd = iden.Max.Time()
			return true
		}

	case *expression.GreaterThanOrEqual:
		if r.isFieldExpression(v.Left(), taskColumnDeliverAt) {
			r.rangeStart = r.getTime(v.Right())
			r.rangeEnd = iden.Max.Time()
			return true
		}

	case *expression.LessThan:
		if r.isFieldExpression(v.Left(), taskColumnDeliverAt) {
			r.rangeStart = iden.Min.Time()
			r.rangeEnd = r.getTime(v.Right())
			return true
		}

	case *expression.LessThanOrEqual:
		if r.isFieldExpression(v.Left(), taskColumnDeliverAt) {
			r.rangeStart = iden.Min.Time()
			r.rangeEnd = r.getTime(v.Right())
			return true
		}

	case *expression.Equals:
		// TODO: This expression type does not work. No results are returned.
		if r.isFieldExpression(v.Left(), taskColumnDeliverAt) {
			r.rangeStart = r.getTime(v.Right())
			r.rangeEnd = r.rangeStart
			return true
		}

	default:
		return false
	}
	return false
}

func (r *taskTableIterator) isFieldExpression(f sql.Expression, field string) bool {
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

func (r *taskTableIterator) getNamespace(f sql.Expression) string {
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

func (r *taskTableIterator) getTime(f sql.Expression) time.Time {
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

func (r *taskTableIterator) detectNamespace(f sql.Expression) bool {
	equal, ok := f.(*expression.Equals)
	if !ok {
		return false
	}

	be := equal.BinaryExpression
	if r.isFieldExpression(be.Left, "namespace") {
		r.namespace = r.getNamespace(be.Right)
		r.namespaceFound = true
		return true
	}
	return false
}

func (r *taskTableIterator) Next() (sql.Row, error) {
	task, err := r.it.Next()
	if err != nil {
		if err != io.EOF {
			telemetry.Logger.Error("encountered error when iterating task range", zap.Error(err))
		}
		return nil, io.EOF
	}
	row := r.rowFromTask(task)

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

// rowFromTask converts a task into a sql row.
func (r *taskTableIterator) rowFromTask(task *proto.Task) sql.Row {
	var lastRetry interface{}
	if task.GetLastRetry() != nil {
		lastRetry = task.GetLastRetry().AsTime()
	}

	return sql.NewRow(
		task.GetNamespace(),
		task.GetId(),
		task.GetScheduleTime().AsTime(),
		lastRetry,
		sql.JSONDocument{Val: task.GetMetadata()},
		task.GetPayload(),
	)
}

func (r *taskTableIterator) Close(context *sql.Context) error {
	return r.it.Close()
}
