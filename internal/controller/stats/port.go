package stats

import (
	"avito_test/internal/domain"
	"context"
	"log/slog"
)

var _ Controller = (*controller)(nil)

type Controller interface {
	GetReviewerAssignmentsStats(ctx context.Context) (AssignmentsStatsOutput, error)
}

type usecase interface {
	GetReviewerAssignmentsStats(ctx context.Context) ([]domain.ReviewerAssignmentStat, error)
}

type controller struct {
	uc     usecase
	logger *slog.Logger
}

func New(uc usecase, logger *slog.Logger) *controller {
	return &controller{
		uc:     uc,
		logger: logger,
	}
}
