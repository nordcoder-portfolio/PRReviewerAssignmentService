package stats

import (
	"avito_test/internal/domain"
	"context"
	"log/slog"
)

type repository interface {
	ListReviewerAssignmentsStats(ctx context.Context) (map[domain.User]int64, error)
}

type usecase struct {
	repo   repository
	logger *slog.Logger
}

func New(repo repository, logger *slog.Logger) *usecase {
	return &usecase{
		repo:   repo,
		logger: logger,
	}
}
