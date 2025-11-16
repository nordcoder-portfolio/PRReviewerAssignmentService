package postgrestest

import (
	"avito_test/internal/domain"
	pgtest "avito_test/internal/repo/postgres/test"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserRepository_GetByID_Success(t *testing.T) {
	ctx, repo := newUserTestRepo(t)

	expected := pgtest.NewUser("u1", "Vasya", pgtest.TeamBackend, true)
	pgtest.InsertUser(t, expected)

	got, err := repo.GetByID(ctx, expected.ID)
	require.NoError(t, err)
	pgtest.RequireUserEqual(t, expected, got)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	ctx, repo := newUserTestRepo(t)

	_, err := repo.GetByID(ctx, "nonexistent")
	require.Error(t, err)

	appErr, ok := domain.AsAppError(err)
	require.True(t, ok, "expected domain.AppError")
	require.Equal(t, domain.CodeNotFound, appErr.Code)
}
