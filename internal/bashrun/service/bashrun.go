package service

import (
	"context"
	"fmt"

	"github.com/PoorMercymain/bashrun/internal/bashrun/domain"
)

var (
	_ domain.BashrunService = (*bashrunService)(nil)
)

type bashrunService struct {
	repo domain.BashrunRepository
}

func New(repo domain.BashrunRepository) *bashrunService {
	return &bashrunService{repo: repo}
}

func (s *bashrunService) Ping(ctx context.Context) error {
	const logPrefix = "service.Ping"

	err := s.repo.Ping(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", logPrefix, err)
	}

	return nil
}