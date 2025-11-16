package postgrestest

import (
	"avito_test/internal/domain"
	pgtest "avito_test/internal/repo/postgres/test"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPRRepository_Create_Success(t *testing.T) {
	ctx, repo := newPRTestRepo(t)

	author := pgtest.MustCreateUser(t, "author1", "Author", pgtest.TeamBackend, true)
	reviewer1 := pgtest.MustCreateUser(t, "rev1", "Reviewer1", pgtest.TeamBackend, true)
	reviewer2 := pgtest.MustCreateUser(t, "rev2", "Reviewer2", pgtest.TeamBackend, true)

	pr := domain.PullRequest{
		ID:                "pr1",
		Name:              "Implement feature X",
		AuthorID:          author.ID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{reviewer1.ID, reviewer2.ID},
		CreatedAt:         pgtest.NowTruncated(),
		MergedAt:          nil,
	}

	err := repo.Create(ctx, pr)
	require.NoError(t, err)

	row := pgtest.GetPRRow(ctx, t, pr.ID)

	stored := domain.PullRequest{
		ID:        row.ID,
		Name:      row.Name,
		AuthorID:  row.AuthorID,
		Status:    domain.PRStatus(row.Status),
		CreatedAt: row.CreatedAt,
		MergedAt:  row.MergedAt,
	}

	pgtest.RequirePREqual(t, pr, stored)

	reviewers := pgtest.GetReviewersForPR(ctx, t, pr.ID)
	require.Equal(t, []string{reviewer1.ID, reviewer2.ID}, reviewers)
}

func TestPRRepository_Create_DuplicateID_ReturnsPRExists(t *testing.T) {
	ctx, repo := newPRTestRepo(t)

	author := pgtest.MustCreateUser(t, "author1", "Author", pgtest.TeamBackend, true)

	pr := domain.PullRequest{
		ID:                "pr1",
		Name:              "Initial PR",
		AuthorID:          author.ID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: nil,
		CreatedAt:         pgtest.NowTruncated(),
		MergedAt:          nil,
	}

	require.NoError(t, repo.Create(ctx, pr))

	err := repo.Create(ctx, pr)
	require.Error(t, err)

	appErr, ok := domain.AsAppError(err)
	require.True(t, ok, "expected domain.AppError from PRExists")
	require.Equal(t, domain.CodePRExists, appErr.Code)

	cnt := pgtest.GetPRCount(ctx, t, pr.ID)
	require.Equal(t, 1, cnt)
}
