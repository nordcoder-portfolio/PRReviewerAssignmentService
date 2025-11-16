package team

import (
	"avito_test/internal/domain"
	"avito_test/internal/repo/postgres"
	"context"
	"log/slog"
)

var _ Usecase = (*usecase)(nil)

type Usecase interface {
	CreateTeam(ctx context.Context, team domain.Team) (domain.Team, error)
	GetTeam(ctx context.Context, name string) (domain.Team, error)
}

type prRepo interface {
	GetByID(ctx context.Context, id string) (domain.PullRequest, error)
	Update(ctx context.Context, pr domain.PullRequest) error
	ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error)
}

type teamRepo interface {
	Create(ctx context.Context, name string) error
	GetByName(ctx context.Context, name string) (domain.Team, error)
}

type userRepo interface {
	UpsertMany(ctx context.Context, teamName string, users []domain.User) error
	SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
}

type transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type reviewerChooser interface {
	Choice(candidates []string, limit int) []string
}

type usecase struct {
	teams   teamRepo
	users   userRepo
	prs     prRepo
	tx      transactor
	chooser reviewerChooser
	logger  *slog.Logger
}

func New(teams teamRepo, users userRepo, prs prRepo, chooser reviewerChooser, tx postgres.Transactor, logger *slog.Logger) *usecase {
	return &usecase{
		teams:   teams,
		users:   users,
		prs:     prs,
		chooser: chooser,
		tx:      tx,
		logger:  logger,
	}
}
