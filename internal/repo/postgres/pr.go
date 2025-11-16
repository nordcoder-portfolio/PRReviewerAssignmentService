//go:build !mockery

package postgres

import (
	"avito_test/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var _ PRRepository = (*prRepository)(nil)

type prRepository struct {
	tx Transactor
}

func NewPRRepository(tx Transactor) *prRepository {
	return &prRepository{tx: tx}
}

const (
	sqlInsertPR = `
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

	sqlInsertPRReviewer = `
		INSERT INTO pr_reviewers (pull_request_id, reviewer_id)
		VALUES ($1, $2);
	`

	sqlSelectPRByID = `
		SELECT
		    pull_request_id,
		    pull_request_name,
		    author_id,
		    status,
		    created_at,
		    merged_at
		FROM pull_requests
		WHERE pull_request_id = $1;
	`

	sqlSelectPRReviewers = `
		SELECT reviewer_id
		FROM pr_reviewers
		WHERE pull_request_id = $1
		ORDER BY reviewer_id;
	`

	sqlUpdatePR = `
		UPDATE pull_requests
		SET status = $2,
		    merged_at = $3
		WHERE pull_request_id = $1;
	`

	sqlDeletePRReviewers = `
		DELETE FROM pr_reviewers
		WHERE pull_request_id = $1;
	`

	sqlListPRsByReviewer = `
	SELECT
	    pr.pull_request_id,
	    pr.pull_request_name,
	    pr.author_id,
	    pr.status,
	    pr.created_at
	FROM pull_requests pr
	JOIN pr_reviewers r
	    ON r.pull_request_id = pr.pull_request_id
	WHERE r.reviewer_id = $1
	ORDER BY pr.created_at, pr.pull_request_id;
`

	sqlAssignmentCount = `
	SELECT
		u.user_id,
		u.username,
		t.name      AS team_name,
		u.is_active,
		COUNT(*)    AS assignments_count
	FROM pr_reviewers r
	JOIN users u ON u.user_id = r.reviewer_id
	JOIN teams t ON t.id = u.team_id
	GROUP BY
		u.user_id,
		u.username,
		t.name,
		u.is_active;
	`
)

func (r *prRepository) Create(ctx context.Context, pr domain.PullRequest) error {
	q := r.tx.Querier(ctx)

	_, err := q.Exec(
		ctx,
		sqlInsertPR,
		pr.ID,
		pr.Name,
		pr.AuthorID,
		string(pr.Status),
		pr.CreatedAt,
		pr.MergedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.PRExists("PR id already exists")
		}
		return fmt.Errorf("insert pull_request %q: %w", pr.ID, err)
	}

	for _, reviewerID := range pr.AssignedReviewers {
		if _, err := q.Exec(ctx, sqlInsertPRReviewer, pr.ID, reviewerID); err != nil {
			return fmt.Errorf("insert reviewer %q for pull_request %q: %w", reviewerID, pr.ID, err)
		}
	}

	return nil
}

func (r *prRepository) GetByID(ctx context.Context, id string) (domain.PullRequest, error) {
	q := r.tx.Querier(ctx)

	var (
		prID      string
		name      string
		authorID  string
		statusStr string
		createdAt time.Time
		mergedAt  *time.Time
	)

	err := q.QueryRow(ctx, sqlSelectPRByID, id).Scan(
		&prID,
		&name,
		&authorID,
		&statusStr,
		&createdAt,
		&mergedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PullRequest{}, domain.NotFound("pull request not found")
		}
		return domain.PullRequest{}, fmt.Errorf("get pull_request %q: %w", id, err)
	}

	rows, err := q.Query(ctx, sqlSelectPRReviewers, prID)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("get reviewers for pull_request %q: %w", prID, err)
	}
	defer rows.Close()

	reviewers := make([]string, 0)

	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return domain.PullRequest{}, fmt.Errorf("scan reviewer for pull_request %q: %w", prID, err)
		}
		reviewers = append(reviewers, reviewerID)
	}

	if err := rows.Err(); err != nil {
		return domain.PullRequest{}, fmt.Errorf("iterate reviewers for pull_request %q: %w", prID, err)
	}

	return domain.PullRequest{
		ID:                prID,
		Name:              name,
		AuthorID:          authorID,
		Status:            domain.PRStatus(statusStr),
		AssignedReviewers: reviewers,
		CreatedAt:         createdAt,
		MergedAt:          mergedAt,
	}, nil
}

func (r *prRepository) Update(ctx context.Context, pr domain.PullRequest) error {
	q := r.tx.Querier(ctx)

	if _, err := q.Exec(
		ctx,
		sqlUpdatePR,
		pr.ID,
		string(pr.Status),
		pr.MergedAt,
	); err != nil {
		return fmt.Errorf("update pull_request %q: %w", pr.ID, err)
	}

	if _, err := q.Exec(ctx, sqlDeletePRReviewers, pr.ID); err != nil {
		return fmt.Errorf("delete reviewers for pull_request %q: %w", pr.ID, err)
	}

	for _, reviewerID := range pr.AssignedReviewers {
		if _, err := q.Exec(ctx, sqlInsertPRReviewer, pr.ID, reviewerID); err != nil {
			return fmt.Errorf("insert reviewer %q for pull_request %q: %w", reviewerID, pr.ID, err)
		}
	}

	return nil
}

func (r *prRepository) ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error) {
	q := r.tx.Querier(ctx)

	rows, err := q.Query(ctx, sqlListPRsByReviewer, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("list pull_requests for reviewer %q: %w", reviewerID, err)
	}
	defer rows.Close()

	result := make([]domain.PullRequestShort, 0)

	for rows.Next() {
		var (
			id        string
			name      string
			authorID  string
			statusStr string
			createdAt time.Time
		)

		if err := rows.Scan(
			&id,
			&name,
			&authorID,
			&statusStr,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan pull_request for reviewer %q: %w", reviewerID, err)
		}

		result = append(result, domain.PullRequestShort{
			ID:       id,
			Name:     name,
			AuthorID: authorID,
			Status:   domain.PRStatus(statusStr),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pull_requests for reviewer %q: %w", reviewerID, err)
	}

	return result, nil
}

func (r *prRepository) ListReviewerAssignmentsStats(ctx context.Context,
) (map[domain.User]int64, error) {
	rows, err := r.tx.Querier(ctx).Query(ctx, sqlAssignmentCount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[domain.User]int64)

	for rows.Next() {
		var (
			u     domain.User
			count int64
		)

		if err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.TeamName,
			&u.IsActive,
			&count,
		); err != nil {
			return nil, err
		}

		result[u] = count
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
