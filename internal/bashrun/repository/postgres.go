package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgres struct {
	*pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *postgres {
	return &postgres{pool}
}

func GetPgxPool(DSN string) (*pgxpool.Pool, error) {
	const logPrefix = "repository.GetPgxPool"

	config, err := pgxpool.ParseConfig(DSN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logPrefix, err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logPrefix, err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logPrefix, err)
	}

	return pool, nil
}

func (p *postgres) WithTransaction(ctx context.Context, txFunc func(pgx.Tx) error) error {
	const logPrefix = "repository.WithTransaction"

	conn, err := p.Pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", logPrefix, err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", logPrefix, err)
	}

	err = txFunc(tx)
	if err != nil {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			resErr := errors.Join(err, rollbackErr)
			return fmt.Errorf("%s: %w", logPrefix, resErr)
		}

		return fmt.Errorf("%s: %w", logPrefix, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", logPrefix, err)
	}

	return nil
}