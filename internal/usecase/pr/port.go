package pr

import (
	"avito_test/internal/domain"
	"context"
	"log/slog"
)

var _ Usecase = (*usecase)(nil)

type Usecase interface {
	Create(ctx context.Context, id, name, authorID string) (domain.PullRequest, error)
	Merge(ctx context.Context, id string) (domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (domain.PullRequest, string, error)
}

type userRepo interface {
	GetByID(ctx context.Context, id string) (domain.User, error)
	ListActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error)
}

type prRepo interface {
	Create(ctx context.Context, pr domain.PullRequest) error
	GetByID(ctx context.Context, id string) (domain.PullRequest, error)
	Update(ctx context.Context, pr domain.PullRequest) error
}

type transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type reviewerChooser interface {
	Choice(candidates []string, limit int) []string
}

type usecase struct {
	users   userRepo
	prs     prRepo
	chooser reviewerChooser
	tx      transactor
	logger  *slog.Logger
}

func New(users userRepo, prs prRepo, chooser reviewerChooser, tx transactor, logger *slog.Logger) *usecase {
	return &usecase{
		users:   users,
		prs:     prs,
		chooser: chooser,
		tx:      tx,
		logger:  logger,
	}
}
