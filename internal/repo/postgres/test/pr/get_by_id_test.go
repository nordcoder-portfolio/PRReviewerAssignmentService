package postgrestest

import (
	"avito_test/internal/domain"
	pgtest "avito_test/internal/repo/postgres/test"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPRRepository_GetByID_SuccessWithReviewers(t *testing.T) {
	ctx, repo := newPRTestRepo(t)

	author := pgtest.MustCreateUser(t, "author1", "Author", pgtest.TeamBackend, true)
	reviewer1 := pgtest.MustCreateUser(t, "rev1", "Reviewer1", pgtest.TeamBackend, true)
	reviewer2 := pgtest.MustCreateUser(t, "rev2", "Reviewer2", pgtest.TeamBackend, true)

	pr := domain.PullRequest{
		ID:        "pr1",
		Name:      "Implement feature X",
		AuthorID:  author.ID,
		Status:    domain.PRStatusOpen,
		CreatedAt: pgtest.NowTruncated(),
		MergedAt:  nil,
	}

	pgtest.InsertPR(t, pr)
	pgtest.InsertReviewer(t, pr.ID, reviewer2.ID)
	pgtest.InsertReviewer(t, pr.ID, reviewer1.ID)

	got, err := repo.GetByID(ctx, pr.ID)
	require.NoError(t, err)

	pgtest.RequirePREqual(t, pr, got)
	require.Equal(t, []string{reviewer1.ID, reviewer2.ID}, got.AssignedReviewers)
}

func TestPRRepository_GetByID_NoReviewers(t *testing.T) {
	ctx, repo := newPRTestRepo(t)

	author := pgtest.MustCreateUser(t, "author2", "Author2", pgtest.TeamBackend, true)

	pr := domain.PullRequest{
		ID:        "pr2",
		Name:      "Lonely PR",
		AuthorID:  author.ID,
		Status:    domain.PRStatusOpen,
		CreatedAt: pgtest.NowTruncated(),
		MergedAt:  nil,
	}

	pgtest.InsertPR(t, pr)

	got, err := repo.GetByID(ctx, pr.ID)
	require.NoError(t, err)

	pgtest.RequirePREqual(t, pr, got)
	require.Empty(t, got.AssignedReviewers)
}

func TestPRRepository_GetByID_NotFound(t *testing.T) {
	ctx, repo := newPRTestRepo(t)

	_, err := repo.GetByID(ctx, "nonexistent")
	require.Error(t, err)

	appErr, ok := domain.AsAppError(err)
	require.True(t, ok, "expected domain.AppError")
	require.Equal(t, domain.CodeNotFound, appErr.Code)
}
