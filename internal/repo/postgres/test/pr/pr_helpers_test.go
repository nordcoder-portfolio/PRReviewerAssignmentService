package postgrestest

import (
	"avito_test/internal/domain"
	"avito_test/internal/repo/postgres"
	pgtest "avito_test/internal/repo/postgres/test"
	"context"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Run(m))
}

type prRepository interface {
	Create(ctx context.Context, pr domain.PullRequest) error
	GetByID(ctx context.Context, id string) (domain.PullRequest, error)
	Update(ctx context.Context, pr domain.PullRequest) error
	ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error)
}

func newPRRepo(t *testing.T) prRepository {
	t.Helper()

	if pgtest.TestPool == nil {
		t.Skip("postgres test pool is not initialized, skipping")
	}

	tx := postgres.NewTransactor(pgtest.TestPool)
	return postgres.NewPRRepository(tx)
}

func newPRTestRepo(t *testing.T) (context.Context, prRepository) {
	t.Helper()

	pgtest.ResetDB(t)
	return t.Context(), newPRRepo(t)
}
