package postgrestest

import (
	"avito_test/internal/domain"
	pgtest "avito_test/internal/repo/postgres/test"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTeamRepository_GetByName_SuccessWithMembers(t *testing.T) {
	ctx, repo := newTeamTestRepo(t)

	backendMembers := []domain.User{
		pgtest.NewUser("u1", "Vasya", pgtest.TeamBackend, true),
		pgtest.NewUser("u2", "Petya", pgtest.TeamBackend, false),
	}
	frontendMembers := []domain.User{
		pgtest.NewUser("u3", "Charlie", pgtest.TeamFrontend, true),
	}

	pgtest.InsertTeamWithMembers(t, pgtest.TeamBackend, backendMembers)
	pgtest.InsertTeamWithMembers(t, pgtest.TeamFrontend, frontendMembers)

	team, err := repo.GetByName(ctx, pgtest.TeamBackend)
	require.NoError(t, err)

	require.Equal(t, pgtest.TeamBackend, team.Name)
	pgtest.RequireTeamMembersMatch(t, backendMembers, team)
}

func TestTeamRepository_GetByName_TeamWithoutMembers(t *testing.T) {
	ctx, repo := newTeamTestRepo(t)

	pgtest.InsertTeamWithMembers(t, pgtest.TeamEmpty, nil)

	team, err := repo.GetByName(ctx, pgtest.TeamEmpty)
	require.NoError(t, err)

	require.Equal(t, pgtest.TeamEmpty, team.Name)
	require.Empty(t, team.Members)
}

func TestTeamRepository_GetByName_NotFound(t *testing.T) {
	ctx, repo := newTeamTestRepo(t)

	_, err := repo.GetByName(ctx, "nonexistent")
	require.Error(t, err)

	appErr, ok := domain.AsAppError(err)
	require.True(t, ok, "expected domain.AppError")
	require.Equal(t, domain.CodeNotFound, appErr.Code)
}
