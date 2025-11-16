package postgrestest

import (
	"avito_test/internal/domain"
	pgtest "avito_test/internal/repo/postgres/test"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserRepository_ListActiveByTeam_FiltersByTeamAndActive(t *testing.T) {
	ctx, repo := newUserTestRepo(t)

	backendUsers := []domain.User{
		pgtest.NewUser("u1", "Vasya", pgtest.TeamBackend, true),
		pgtest.NewUser("u2", "Petya", pgtest.TeamBackend, false),
		pgtest.NewUser("u3", "Charlie", pgtest.TeamBackend, true),
	}
	frontendUsers := []domain.User{
		pgtest.NewUser("u4", "Dave", pgtest.TeamFrontend, true),
	}

	for _, u := range append(backendUsers, frontendUsers...) {
		pgtest.InsertUser(t, u)
	}

	got, err := repo.ListActiveByTeam(ctx, pgtest.TeamBackend)
	require.NoError(t, err)

	expected := []domain.User{
		backendUsers[0],
		backendUsers[2],
	}

	pgtest.RequireUsersMatch(t, expected, got)
}

func TestUserRepository_ListActiveByTeam_NoActiveUsers(t *testing.T) {
	ctx, repo := newUserTestRepo(t)

	users := []domain.User{
		pgtest.NewUser("u1", "Vasya", pgtest.TeamBackend, false),
		pgtest.NewUser("u2", "Petya", pgtest.TeamBackend, false),
	}

	for _, u := range users {
		pgtest.InsertUser(t, u)
	}

	got, err := repo.ListActiveByTeam(ctx, pgtest.TeamBackend)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestUserRepository_ListActiveByTeam_TeamNotExists(t *testing.T) {
	ctx, repo := newUserTestRepo(t)

	pgtest.InsertUser(t, pgtest.NewUser("u1", "Vasya", "other", true))

	got, err := repo.ListActiveByTeam(ctx, pgtest.TeamBackend)
	require.NoError(t, err)
	require.Empty(t, got)
}
