package db

import "github.com/jackc/pgx/v5/pgxpool"

type Database struct {
	Idempotency Idempotency
}

func InitDb(pool *pgxpool.Pool) *Database {
	idempotency := Idempotency{
		pool: pool,
	}

	return &Database{
		Idempotency: idempotency,
	}
}
