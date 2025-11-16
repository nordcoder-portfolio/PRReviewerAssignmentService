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

type userRepository interface {
	UpsertMany(ctx context.Context, teamName string, users []domain.User) error
	GetByID(ctx context.Context, userID string) (domain.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
	ListActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error)
}

func newUserRepo(t *testing.T) userRepository {
	t.Helper()

	if pgtest.TestPool == nil {
		t.Skip("postgres test pool is not initialized, skipping")
	}

	tx := postgres.NewTransactor(pgtest.TestPool)
	repo := postgres.NewUserRepository(tx)

	var r userRepository = repo
	return r
}

func newUserTestRepo(t *testing.T) (context.Context, userRepository) {
	t.Helper()

	pgtest.ResetDB(t)
	return t.Context(), newUserRepo(t)
}
