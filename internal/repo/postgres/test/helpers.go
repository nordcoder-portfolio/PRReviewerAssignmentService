package postgrestest

import (
	"avito_test/internal/domain"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	TeamBackend  = "backend"
	TeamFrontend = "frontend"
	TeamEmpty    = "empty"
)

func NowTruncated() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

func NewUser(id, username, team string, isActive bool) domain.User {
	return domain.User{
		ID:       id,
		Username: username,
		TeamName: team,
		IsActive: isActive,
	}
}

func MustCreateUser(t *testing.T, id, username, team string, isActive bool) domain.User {
	t.Helper()
	u := NewUser(id, username, team, isActive)
	InsertUser(t, u)
	return u
}

func InsertTeam(t *testing.T, name string) int64 {
	t.Helper()

	ctx := t.Context()

	const upsertTeam = `
		INSERT INTO teams (name)
		VALUES ($1)
		ON CONFLICT (name) DO NOTHING;
	`
	_, err := TestPool.Exec(ctx, upsertTeam, name)
	require.NoError(t, err)

	const selectID = `
		SELECT id
		FROM teams
		WHERE name = $1;
	`
	var id int64
	err = TestPool.QueryRow(ctx, selectID, name).Scan(&id)
	require.NoError(t, err)

	return id
}

func InsertUser(t *testing.T, u domain.User) {
	t.Helper()

	ctx := t.Context()

	teamID := InsertTeam(t, u.TeamName)

	const insertUserQuery = `
		INSERT INTO users (user_id, username, team_id, is_active)
		VALUES ($1, $2, $3, $4);
	`

	_, err := TestPool.Exec(
		ctx,
		insertUserQuery,
		u.ID,
		u.Username,
		teamID,
		u.IsActive,
	)
	require.NoError(t, err)
}

func GetAllTeamNames(ctx context.Context, t *testing.T) []string {
	t.Helper()

	const q = `
		SELECT name
		FROM teams
		ORDER BY id;
	`

	rows, err := TestPool.Query(ctx, q)
	require.NoError(t, err)
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		names = append(names, name)
	}
	require.NoError(t, rows.Err())

	return names
}

func InsertTeamWithMembers(t *testing.T, teamName string, members []domain.User) {
	t.Helper()

	ctx := t.Context()

	teamID := InsertTeam(t, teamName)

	if len(members) == 0 {
		return
	}

	const insertUserQuery = `
		INSERT INTO users (user_id, username, team_id, is_active)
		VALUES ($1, $2, $3, $4);
	`

	for _, m := range members {
		_, err := TestPool.Exec(
			ctx,
			insertUserQuery,
			m.ID,
			m.Username,
			teamID,
			m.IsActive,
		)
		require.NoError(t, err)
	}
}

func RequireTeamMembersMatch(t *testing.T, expected []domain.User, team domain.Team) {
	t.Helper()

	require.Len(t, team.Members, len(expected))
	require.ElementsMatch(t, expected, team.Members)
}

type UserRow struct {
	ID       string
	Username string
	TeamName string
	IsActive bool
}

func GetUserRow(ctx context.Context, t *testing.T, userID string) UserRow {
	t.Helper()

	const q = `
		SELECT u.user_id, u.username, t.name, u.is_active
		FROM users u
		JOIN teams t ON u.team_id = t.id
		WHERE u.user_id = $1;
	`

	var row UserRow
	dbRow := TestPool.QueryRow(ctx, q, userID)
	require.NoError(t, dbRow.Scan(
		&row.ID,
		&row.Username,
		&row.TeamName,
		&row.IsActive,
	))

	return row
}

type SimpleUserRow struct {
	ID       string
	Username string
	IsActive bool
}

func ListSimpleUsers(ctx context.Context, t *testing.T) []SimpleUserRow {
	t.Helper()

	const q = `
		SELECT user_id, username, is_active
		FROM users
		ORDER BY user_id;
	`

	rows, err := TestPool.Query(ctx, q)
	require.NoError(t, err)
	defer rows.Close()

	var users []SimpleUserRow
	for rows.Next() {
		var u SimpleUserRow
		require.NoError(t, rows.Scan(&u.ID, &u.Username, &u.IsActive))
		users = append(users, u)
	}
	require.NoError(t, rows.Err())

	return users
}

func RequireUserEqual(t *testing.T, want, got domain.User) {
	t.Helper()
	require.Equal(t, want, got)
}

func RequireUsersMatch(t *testing.T, want, got []domain.User) {
	t.Helper()
	require.Len(t, got, len(want))
	require.ElementsMatch(t, want, got)
}

type PRRow struct {
	ID        string
	Name      string
	AuthorID  string
	Status    string
	CreatedAt time.Time
	MergedAt  *time.Time
}

func InsertPR(t *testing.T, pr domain.PullRequest) {
	t.Helper()

	ctx := t.Context()

	const insertPRQuery = `
		INSERT INTO pull_requests (
		    pull_request_id,
		    pull_request_name,
		    author_id,
		    status,
		    created_at,
		    merged_at
		)
		VALUES ($1, $2, $3, $4, $5, $6);
	`

	_, err := TestPool.Exec(
		ctx,
		insertPRQuery,
		pr.ID,
		pr.Name,
		pr.AuthorID,
		string(pr.Status),
		pr.CreatedAt,
		pr.MergedAt,
	)
	require.NoError(t, err)
}

func InsertReviewer(t *testing.T, prID, reviewerID string) {
	t.Helper()

	ctx := t.Context()
	const insertReviewerQuery = `
		INSERT INTO pr_reviewers (pull_request_id, reviewer_id)
		VALUES ($1, $2);
	`

	_, err := TestPool.Exec(ctx, insertReviewerQuery, prID, reviewerID)
	require.NoError(t, err)
}

func GetReviewersForPR(ctx context.Context, t *testing.T, prID string) []string {
	t.Helper()

	const reviewersQuery = `
		SELECT reviewer_id
		FROM pr_reviewers
		WHERE pull_request_id = $1
		ORDER BY reviewer_id;
	`

	rows, err := TestPool.Query(ctx, reviewersQuery, prID)
	require.NoError(t, err)
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var rid string
		require.NoError(t, rows.Scan(&rid))
		reviewers = append(reviewers, rid)
	}
	require.NoError(t, rows.Err())

	return reviewers
}

func GetPRRow(ctx context.Context, t *testing.T, prID string) PRRow {
	t.Helper()

	const prQuery = `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1;
	`

	var row PRRow
	dbRow := TestPool.QueryRow(ctx, prQuery, prID)
	require.NoError(t, dbRow.Scan(
		&row.ID,
		&row.Name,
		&row.AuthorID,
		&row.Status,
		&row.CreatedAt,
		&row.MergedAt,
	))

	return row
}

func GetPRCount(ctx context.Context, t *testing.T, prID string) int {
	t.Helper()

	const countQuery = `
		SELECT COUNT(*)
		FROM pull_requests
		WHERE pull_request_id = $1;
	`

	var cnt int
	row := TestPool.QueryRow(ctx, countQuery, prID)
	require.NoError(t, row.Scan(&cnt))

	return cnt
}

func RequirePREqual(t *testing.T, want, got domain.PullRequest) {
	t.Helper()

	require.Equal(t, want.ID, got.ID)
	require.Equal(t, want.Name, got.Name)
	require.Equal(t, want.AuthorID, got.AuthorID)
	require.Equal(t, want.Status, got.Status)
	require.WithinDuration(t, want.CreatedAt, got.CreatedAt, time.Second)

	if want.MergedAt == nil {
		require.Nil(t, got.MergedAt)
	} else {
		require.NotNil(t, got.MergedAt)
		require.WithinDuration(t, *want.MergedAt, *got.MergedAt, time.Second)
	}
}

func RequirePRShortEqual(t *testing.T, want domain.PullRequest, got domain.PullRequestShort) {
	t.Helper()

	require.Equal(t, want.ID, got.ID)
	require.Equal(t, want.Name, got.Name)
	require.Equal(t, want.AuthorID, got.AuthorID)
	require.Equal(t, want.Status, got.Status)
}
