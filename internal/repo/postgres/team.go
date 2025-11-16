//go:build !mockery

package postgres

import (
	"avito_test/internal/domain"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var _ TeamRepository = (*teamRepository)(nil)

type teamRepository struct {
	tx Transactor
}

func NewTeamRepository(tx Transactor) *teamRepository {
	return &teamRepository{tx: tx}
}

const (
	pgCodeUniqueViolation = "23505"

	sqlInsertTeam = `
		INSERT INTO teams (name)
		VALUES ($1);
	`

	sqlSelectTeamByName = `
		SELECT id, name
		FROM teams
		WHERE name = $1;
	`

	sqlSelectTeamMembers = `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_id = $1
		ORDER BY user_id;
	`
)

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == pgCodeUniqueViolation
}

func (r *teamRepository) Create(ctx context.Context, name string) error {
	q := r.tx.Querier(ctx)

	if _, err := q.Exec(ctx, sqlInsertTeam, name); err != nil {
		if isUniqueViolation(err) {
			return domain.TeamExists("team_name already exists")
		}
		return fmt.Errorf("create team %q: %w", name, err)
	}

	return nil
}

func (r *teamRepository) GetByName(ctx context.Context, name string) (domain.Team, error) {
	q := r.tx.Querier(ctx)

	var (
		teamID   int64
		teamName string
	)

	if err := q.QueryRow(ctx, sqlSelectTeamByName, name).Scan(&teamID, &teamName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Team{}, domain.NotFound("team not found")
		}
		return domain.Team{}, fmt.Errorf("get team %q: %w", name, err)
	}

	rows, err := q.Query(ctx, sqlSelectTeamMembers, teamID)
	if err != nil {
		return domain.Team{}, fmt.Errorf("get team %q members: %w", teamName, err)
	}
	defer rows.Close()

	members := make([]domain.User, 0)

	for rows.Next() {
		var (
			userID   string
			username string
			isActive bool
		)

		if err := rows.Scan(&userID, &username, &isActive); err != nil {
			return domain.Team{}, fmt.Errorf("scan team %q member: %w", teamName, err)
		}

		members = append(members, domain.User{
			ID:       userID,
			Username: username,
			TeamName: teamName,
			IsActive: isActive,
		})
	}

	if err := rows.Err(); err != nil {
		return domain.Team{}, fmt.Errorf("iterate team %q members: %w", teamName, err)
	}

	return domain.Team{
		Name:    teamName,
		Members: members,
	}, nil
}
