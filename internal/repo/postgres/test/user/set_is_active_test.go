package postgrestest

import (
	"avito_test/internal/domain"
	pgtest "avito_test/internal/repo/postgres/test"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserRepository_SetIsActive_DeactivateUser(t *testing.T) {
	ctx, repo := newUserTestRepo(t)

	initial := pgtest.NewUser("u1", "Vasya", pgtest.TeamBackend, true)
	pgtest.InsertUser(t, initial)

	updated, err := repo.SetIsActive(ctx, initial.ID, false)
	require.NoError(t, err)

	pgtest.RequireUserEqual(t, domain.User{
		ID:       initial.ID,
		Username: initial.Username,
		TeamName: initial.TeamName,
		IsActive: false,
	}, updated)

	row := pgtest.GetUserRow(ctx, t, initial.ID)
	require.Equal(t, initial.ID, row.ID)
	require.Equal(t, initial.Username, row.Username)
	require.Equal(t, initial.TeamName, row.TeamName)
	require.False(t, row.IsActive)
}

func TestUserRepository_SetIsActive_ActivateUser(t *testing.T) {
	ctx, repo := newUserTestRepo(t)

	initial := pgtest.NewUser("u2", "Petya", pgtest.TeamBackend, false)
	pgtest.InsertUser(t, initial)

	updated, err := repo.SetIsActive(ctx, initial.ID, true)
	require.NoError(t, err)

	pgtest.RequireUserEqual(t, domain.User{
		ID:       initial.ID,
		Username: initial.Username,
		TeamName: initial.TeamName,
		IsActive: true,
	}, updated)

	row := pgtest.GetUserRow(ctx, t, initial.ID)
	require.Equal(t, initial.ID, row.ID)
	require.Equal(t, initial.Username, row.Username)
	require.Equal(t, initial.TeamName, row.TeamName)
	require.True(t, row.IsActive)
}

func TestUserRepository_SetIsActive_NotFound(t *testing.T) {
	ctx, repo := newUserTestRepo(t)

	_, err := repo.SetIsActive(ctx, "nonexistent", true)
	require.Error(t, err)

	appErr, ok := domain.AsAppError(err)
	require.True(t, ok, "expected domain.AppError")
	require.Equal(t, domain.CodeNotFound, appErr.Code)
}
