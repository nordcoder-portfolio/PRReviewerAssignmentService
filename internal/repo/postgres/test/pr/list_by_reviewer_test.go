package postgrestest

import (
	"avito_test/internal/domain"
	pgtest "avito_test/internal/repo/postgres/test"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPRRepository_ListByReviewer_ReturnsSortedPRsForReviewer(t *testing.T) {
	ctx, repo := newPRTestRepo(t)

	author := pgtest.MustCreateUser(t, "author1", "Author", pgtest.TeamBackend, true)
	reviewer1 := pgtest.MustCreateUser(t, "rev1", "Reviewer1", pgtest.TeamBackend, true)
	reviewer2 := pgtest.MustCreateUser(t, "rev2", "Reviewer2", pgtest.TeamBackend, true)

	base := pgtest.NowTruncated().Add(-time.Hour)

	pr1 := domain.PullRequest{
		ID:        "pr1",
		Name:      "PR 1",
		AuthorID:  author.ID,
		Status:    domain.PRStatusOpen,
		CreatedAt: base,
		MergedAt:  nil,
	}
	pr2 := domain.PullRequest{
		ID:        "pr2",
		Name:      "PR 2",
		AuthorID:  author.ID,
		Status:    domain.PRStatusOpen,
		CreatedAt: base.Add(10 * time.Minute),
		MergedAt:  nil,
	}
	pr3 := domain.PullRequest{
		ID:        "pr3",
		Name:      "PR 3",
		AuthorID:  author.ID,
		Status:    domain.PRStatusOpen,
		CreatedAt: base.Add(20 * time.Minute),
		MergedAt:  nil,
	}

	pgtest.InsertPR(t, pr1)
	pgtest.InsertPR(t, pr2)
	pgtest.InsertPR(t, pr3)

	pgtest.InsertReviewer(t, pr1.ID, reviewer1.ID)
	pgtest.InsertReviewer(t, pr2.ID, reviewer2.ID)
	pgtest.InsertReviewer(t, pr3.ID, reviewer1.ID)

	prs, err := repo.ListByReviewer(ctx, reviewer1.ID)
	require.NoError(t, err)
	require.Len(t, prs, 2)

	pgtest.RequirePRShortEqual(t, pr1, prs[0])
	pgtest.RequirePRShortEqual(t, pr3, prs[1])
}

func TestPRRepository_ListByReviewer_NoPRsForReviewer(t *testing.T) {
	ctx, repo := newPRTestRepo(t)

	author := pgtest.MustCreateUser(t, "authorX", "AuthorX", pgtest.TeamBackend, true)
	otherReviewer := pgtest.MustCreateUser(t, "rev_other", "OtherReviewer", pgtest.TeamBackend, true)

	pr := domain.PullRequest{
		ID:        "prX",
		Name:      "Some PR",
		AuthorID:  author.ID,
		Status:    domain.PRStatusOpen,
		CreatedAt: pgtest.NowTruncated().Add(-30 * time.Minute),
		MergedAt:  nil,
	}

	pgtest.InsertPR(t, pr)
	pgtest.InsertReviewer(t, pr.ID, otherReviewer.ID)

	prs, err := repo.ListByReviewer(ctx, "rev_no_prs")
	require.NoError(t, err)
	require.Empty(t, prs)
}
