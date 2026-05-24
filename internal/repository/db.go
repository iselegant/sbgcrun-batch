package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DB struct {
	*sqlx.DB
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// BeginTx starts a new transaction
func (db *DB) BeginTx() (*sqlx.Tx, error) {
	return db.DB.Beginx()
}

// QueryContext wraps sqlx.DB.QueryxContext
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return db.DB.QueryxContext(ctx, query, args...)
}

// QueryxContext wraps sqlx.DB.QueryxContext
func (db *DB) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return db.DB.QueryxContext(ctx, query, args...)
}

// ExecContext wraps sqlx.DB.ExecContext
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}

// QueryRowContext wraps sqlx.DB.QueryRowContext
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(ctx, query, args...)
}
