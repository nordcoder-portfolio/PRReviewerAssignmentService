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

type teamRepository interface {
	Create(ctx context.Context, name string) error
	GetByName(ctx context.Context, name string) (domain.Team, error)
}

func newTeamRepo(t *testing.T) teamRepository {
	t.Helper()

	if pgtest.TestPool == nil {
		t.Skip("postgres test pool is not initialized, skipping")
	}

	tx := postgres.NewTransactor(pgtest.TestPool)
	return postgres.NewTeamRepository(tx)
}

func newTeamTestRepo(t *testing.T) (context.Context, teamRepository) {
	t.Helper()

	pgtest.ResetDB(t)
	return t.Context(), newTeamRepo(t)
}
