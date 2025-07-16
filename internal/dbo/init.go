package dbo

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io"

	migrate "github.com/rubenv/sql-migrate"
	_ "modernc.org/sqlite"
)

//go:generate sqlc generate
//go:embed migrations
var migrations embed.FS

func NewFromFile(file string) (*Queries, error) {
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&cache=shared&_pragma=foreign_keys(0)", file) // migrate without FK
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	_, err = migrate.Exec(db, "sqlite3", &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrations,
		Root:       "migrations",
	}, migrate.Up)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate db: %w", err)
	}
	_ = db.Close()

	// enable back FK
	dsn = fmt.Sprintf("file:%s?_journal_mode=WAL&cache=shared&_pragma=foreign_keys(1)", file)
	db, err = sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)
	return New(db), nil
}

func (q *Queries) Close() error {
	if v, ok := q.db.(io.Closer); ok && v != nil {
		return v.Close()
	}
	return nil
}

func (q *Queries) Transaction(ctx context.Context, tx func(q *Queries) error) error {
	type transactor interface {
		BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	}

	v, ok := q.db.(transactor)
	if !ok {
		return fmt.Errorf("db does not support transactions")
	}
	t, err := v.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	res := tx(q.WithTx(t))
	if res != nil {
		return errors.Join(res, t.Rollback())
	}

	return t.Commit()
}
