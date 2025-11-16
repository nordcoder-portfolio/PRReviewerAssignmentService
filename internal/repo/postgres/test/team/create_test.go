package postgrestest

import (
	"avito_test/internal/domain"
	pgtest "avito_test/internal/repo/postgres/test"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTeamRepository_Create_Success(t *testing.T) {
	ctx, repo := newTeamTestRepo(t)

	const teamName = pgtest.TeamBackend

	err := repo.Create(ctx, teamName)
	require.NoError(t, err)

	names := pgtest.GetAllTeamNames(ctx, t)
	require.Len(t, names, 1)
	require.Equal(t, teamName, names[0])
}

func TestTeamRepository_Create_DuplicateName_ReturnsTeamExists(t *testing.T) {
	ctx, repo := newTeamTestRepo(t)

	const teamName = pgtest.TeamBackend

	require.NoError(t, repo.Create(ctx, teamName))

	err := repo.Create(ctx, teamName)
	require.Error(t, err)

	appErr, ok := domain.AsAppError(err)
	require.True(t, ok, "expected domain.AppError from TeamExists")
	require.Equal(t, domain.CodeTeamExists, appErr.Code)

	names := pgtest.GetAllTeamNames(ctx, t)
	require.Len(t, names, 1)
	require.Equal(t, teamName, names[0])
}
