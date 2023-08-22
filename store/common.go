package store

import (
	"context"
	"database/sql"
)

// queryable allows to pass *sql.DB or *sql.Tx interchangeably to the consuming methods.
type queryable interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
