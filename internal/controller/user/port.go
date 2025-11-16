package user

import (
	"avito_test/internal/domain"
	"context"
	"log/slog"
)

var _ Controller = (*controller)(nil)

type Controller interface {
	SetIsActive(ctx context.Context, in SetIsActiveInput) (SetIsActiveOutput, error)
	GetReviewPRs(ctx context.Context, in GetReviewInput) (GetReviewOutput, error)
}

type usecase interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
	GetReviewPRs(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
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
