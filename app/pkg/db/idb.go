package db

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (bun.Tx, error) {
	return d.PrimaryConn().BeginTx(ctx, opts)
}

func (d *DB) Dialect() schema.Dialect {
	return d.PrimaryConn().Dialect()
}

func (d *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.PrimaryConn().ExecContext(ctx, query, args...)
}

func (d *DB) NewAddColumn() *bun.AddColumnQuery {
	return d.PrimaryConn().NewAddColumn()
}

func (d *DB) NewCreateIndex() *bun.CreateIndexQuery {
	return d.PrimaryConn().NewCreateIndex()
}

func (d *DB) NewCreateTable() *bun.CreateTableQuery {
	return d.PrimaryConn().NewCreateTable()
}

func (d *DB) NewDelete() *bun.DeleteQuery {
	return d.PrimaryConn().NewDelete()
}

func (d *DB) NewDropColumn() *bun.DropColumnQuery {
	return d.PrimaryConn().NewDropColumn()
}

func (d *DB) NewDropIndex() *bun.DropIndexQuery {
	return d.PrimaryConn().NewDropIndex()
}

func (d *DB) NewDropTable() *bun.DropTableQuery {
	return d.PrimaryConn().NewDropTable()
}

func (d *DB) NewInsert() *bun.InsertQuery {
	return d.PrimaryConn().NewInsert()
}

func (d *DB) NewMerge() *bun.MergeQuery {
	return d.PrimaryConn().NewMerge()
}

func (d *DB) NewRaw(query string, args ...interface{}) *bun.RawQuery {
	return d.PrimaryConn().NewRaw(query, args...)
}

func (d *DB) NewSelect() *bun.SelectQuery {
	return d.PrimaryConn().NewSelect()
}

func (d *DB) NewTruncateTable() *bun.TruncateTableQuery {
	return d.PrimaryConn().NewTruncateTable()
}

func (d *DB) NewUpdate() *bun.UpdateQuery {
	return d.PrimaryConn().NewUpdate()
}

func (d *DB) NewValues(model interface{}) *bun.ValuesQuery {
	return d.PrimaryConn().NewValues(model)
}

func (d *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.PrimaryConn().QueryContext(ctx, query, args...)
}

func (d *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.PrimaryConn().QueryRowContext(ctx, query, args...)
}

func (d *DB) RunInTx(ctx context.Context, opts *sql.TxOptions, f func(ctx context.Context, tx bun.Tx) error) error {
	return d.PrimaryConn().RunInTx(ctx, opts, f)
}

func (d *DB) ReplicaNewSelect() *bun.SelectQuery {
	return d.ReplicaConn().NewSelect()
}

func (d *DB) ReplicaQueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.ReplicaConn().QueryContext(ctx, query, args...)
}

func (d *DB) ReplicaQueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.ReplicaConn().QueryRowContext(ctx, query, args...)
}

var _ bun.IDB = (*DB)(nil)
