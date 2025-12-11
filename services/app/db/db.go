package db

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNoRows is returned when a query expects a row but finds none.
// Callers should check for this specific error to handle "not found"
// cases.
var ErrNoRows = errors.New("db: no rows in result set")

// Database is a facade around the database queries provided by sqlc.
// Queries are grouped by their functionality.
//
// Database is intended to be used in main.go, and passed to the API
// handlers.
type Database struct {
	Idempotency Idempotency
}

// Return a database instance for a PostgreSQL pool connection
func InitDb(pool *pgxpool.Pool) *Database {
	idempotency := Idempotency{
		pool: pool,
	}

	return &Database{
		Idempotency: idempotency,
	}
}
