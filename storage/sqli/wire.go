package sqli

import (
	"fmt"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/auth"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/information_schema"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"io"
)

// TODO: Refine this implementation.
// SQL support is still pretty fragile. We need to increase the robustness
// of the namespace/timestamp indexing and also add functional tests.

type WireConfig struct {
	Port  int
	Tasks *storage.TaskStore
}

func StartWireListener(config WireConfig) error {
	db := NewDatabase("z4")
	db.AddTable(newTaskTable(config.Tasks))

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

// noOpPartitionIterator is a no-op iterator; we don't support partitions
type noOpPartitionIterator struct {
	i int
}

func (p *noOpPartitionIterator) Close(context *sql.Context) error {
	return nil
}

func (p *noOpPartitionIterator) Next() (sql.Partition, error) {
	if p.i == 0 {
		p.i++
		return &noOpPartition{}, nil
	}
	return nil, io.EOF
}

// noOpPartition is a no-op partition; we don't support partitions
type noOpPartition struct {
}

func (p *noOpPartition) Key() []byte {
	return nil
}
