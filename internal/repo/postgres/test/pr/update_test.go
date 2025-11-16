package postgrestest

import (
	"avito_test/internal/domain"
	pgtest "avito_test/internal/repo/postgres/test"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPRRepository_Update_ChangesStatusAndReviewers(t *testing.T) {
	ctx, repo := newPRTestRepo(t)

	author := pgtest.MustCreateUser(t, "author1", "Author", pgtest.TeamBackend, true)
	_ = pgtest.MustCreateUser(t, "rev1", "Reviewer1", pgtest.TeamBackend, true)
	_ = pgtest.MustCreateUser(t, "rev2", "Reviewer2", pgtest.TeamBackend, true)
	newReviewer := pgtest.MustCreateUser(t, "rev3", "Reviewer3", pgtest.TeamBackend, true)

	pr := domain.PullRequest{
		ID:        "pr1",
		Name:      "Initial PR",
		AuthorID:  author.ID,
		Status:    domain.PRStatusOpen,
		CreatedAt: pgtest.NowTruncated().Add(-time.Hour),
		MergedAt:  nil,
	}

	pgtest.InsertPR(t, pr)
	pgtest.InsertReviewer(t, pr.ID, "rev1")
	pgtest.InsertReviewer(t, pr.ID, "rev2")

	mergedAt := pgtest.NowTruncated()

	err := repo.Update(ctx, domain.PullRequest{
		ID:                pr.ID,
		Status:            domain.PRStatusMerged,
		MergedAt:          &mergedAt,
		AssignedReviewers: []string{newReviewer.ID},
	})
	require.NoError(t, err)

	row := pgtest.GetPRRow(ctx, t, pr.ID)

	expected := domain.PullRequest{
		ID:        pr.ID,
		Name:      pr.Name,
		AuthorID:  pr.AuthorID,
		Status:    domain.PRStatusMerged,
		CreatedAt: pr.CreatedAt,
		MergedAt:  &mergedAt,
	}

	stored := domain.PullRequest{
		ID:        row.ID,
		Name:      row.Name,
		AuthorID:  row.AuthorID,
		Status:    domain.PRStatus(row.Status),
		CreatedAt: row.CreatedAt,
		MergedAt:  row.MergedAt,
	}

	pgtest.RequirePREqual(t, expected, stored)

	reviewers := pgtest.GetReviewersForPR(ctx, t, pr.ID)
	require.Equal(t, []string{newReviewer.ID}, reviewers)
}

func TestPRRepository_Update_ClearsReviewersWhenEmpty(t *testing.T) {
	ctx, repo := newPRTestRepo(t)

	author := pgtest.MustCreateUser(t, "author2", "Author2", pgtest.TeamBackend, true)
	reviewer := pgtest.MustCreateUser(t, "rev1", "Reviewer1", pgtest.TeamBackend, true)

	pr := domain.PullRequest{
		ID:        "pr2",
		Name:      "PR with reviewers",
		AuthorID:  author.ID,
		Status:    domain.PRStatusOpen,
		CreatedAt: pgtest.NowTruncated().Add(-2 * time.Hour),
		MergedAt:  nil,
	}

	pgtest.InsertPR(t, pr)
	pgtest.InsertReviewer(t, pr.ID, reviewer.ID)

	err := repo.Update(ctx, domain.PullRequest{
		ID:                pr.ID,
		Status:            domain.PRStatusOpen,
		MergedAt:          nil,
		AssignedReviewers: nil,
	})
	require.NoError(t, err)

	reviewers := pgtest.GetReviewersForPR(ctx, t, pr.ID)
	require.Empty(t, reviewers)
}
