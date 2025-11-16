package postgrestest

import (
	"avito_test/internal/domain"
	pgtest "avito_test/internal/repo/postgres/test"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserRepository_UpsertMany_InsertNewUsers(t *testing.T) {
	ctx, repo := newUserTestRepo(t)

	teamName := pgtest.TeamBackend
	pgtest.InsertTeam(t, teamName)

	users := []domain.User{
		{ID: "u1", Username: "Vasya", IsActive: true},
		{ID: "u2", Username: "Petya", IsActive: false},
	}

	err := repo.UpsertMany(ctx, teamName, users)
	require.NoError(t, err)

	got := pgtest.ListSimpleUsers(ctx, t)
	require.Len(t, got, 2)

	require.Equal(t, pgtest.SimpleUserRow{ID: "u1", Username: "Vasya", IsActive: true}, got[0])
	require.Equal(t, pgtest.SimpleUserRow{ID: "u2", Username: "Petya", IsActive: false}, got[1])
}

func TestUserRepository_UpsertMany_UpdateExistingUsers(t *testing.T) {
	ctx, repo := newUserTestRepo(t)

	teamName := pgtest.TeamBackend
	pgtest.InsertTeam(t, teamName)

	initial := []domain.User{
		{ID: "u1", Username: "Vasya", IsActive: true},
	}

	updated := []domain.User{
		{ID: "u1", Username: "Vasya2", IsActive: false},
	}

	require.NoError(t, repo.UpsertMany(ctx, teamName, initial))
	require.NoError(t, repo.UpsertMany(ctx, teamName, updated))

	got := pgtest.ListSimpleUsers(ctx, t)
	require.Len(t, got, 1)

	require.Equal(t, pgtest.SimpleUserRow{
		ID:       "u1",
		Username: "Vasya2",
		IsActive: false,
	}, got[0])
}

func TestUserRepository_UpsertMany_TeamNotFound(t *testing.T) {
	ctx, repo := newUserTestRepo(t)

	users := []domain.User{
		{ID: "u1", Username: "Vasya", IsActive: true},
	}

	err := repo.UpsertMany(ctx, "nonexistent", users)
	require.Error(t, err)

	appErr, ok := domain.AsAppError(err)
	require.True(t, ok, "expected domain.AppError")
	require.Equal(t, domain.CodeNotFound, appErr.Code)
}
