package repository

import (
	"context"
	"fmt"

	"github.com/PoorMercymain/bashrun/internal/bashrun/domain"
)

var (
	_ domain.BashrunRepository = (*bashrunRepository)(nil)
)

type bashrunRepository struct {
	db *postgres
}

func New(db *postgres) *bashrunRepository {
	return &bashrunRepository{db: db}
}

func (r *bashrunRepository) Ping(ctx context.Context) error {
	const logPrefix = "repository.Ping"
	err := r.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", logPrefix, err)
	}

	return nil
}
