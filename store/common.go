package store

import (
	"context"
	"database/sql"
)

type queryable interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
