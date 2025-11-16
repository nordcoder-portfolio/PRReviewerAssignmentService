package postgres

import (
	"avito_test/internal/domain"
	"context"
)

type UserRepository interface {
	UpsertMany(ctx context.Context, teamName string, users []domain.User) error
	SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
	GetByID(ctx context.Context, userID string) (domain.User, error)
	ListActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error)
}

type TeamRepository interface {
	Create(ctx context.Context, name string) error
	GetByName(ctx context.Context, name string) (domain.Team, error)
}

type PRRepository interface {
	Create(ctx context.Context, pr domain.PullRequest) error
	GetByID(ctx context.Context, id string) (domain.PullRequest, error)
	Update(ctx context.Context, pr domain.PullRequest) error
	ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error)
	ListReviewerAssignmentsStats(ctx context.Context) (map[domain.User]int64, error)
}
