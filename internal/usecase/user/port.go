package user

import (
	"avito_test/internal/domain"
	"context"
	"log/slog"
)

var _ Usecase = (*usecase)(nil)

type Usecase interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
	GetReviewPRs(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
}

type userRepo interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
}

type prRepo interface {
	ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error)
}

type transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type usecase struct {
	users  userRepo
	prs    prRepo
	tx     transactor
	logger *slog.Logger
}

func New(users userRepo, prs prRepo, tx transactor, logger *slog.Logger) *usecase {
	return &usecase{
		users:  users,
		prs:    prs,
		tx:     tx,
		logger: logger,
	}
}
