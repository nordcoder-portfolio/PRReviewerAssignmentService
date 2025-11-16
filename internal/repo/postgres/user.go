//go:build !mockery

package postgres

import (
	"avito_test/internal/domain"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var _ UserRepository = (*userRepository)(nil)

type userRepository struct {
	tx Transactor
}

func NewUserRepository(tx Transactor) *userRepository {
	return &userRepository{tx: tx}
}

const (
	sqlSelectTeamIDByName = `
		SELECT id
		FROM teams
		WHERE name = $1;
	`

	sqlUpsertUser = `
		INSERT INTO users (user_id, username, team_id, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET
			username  = EXCLUDED.username,
			team_id   = EXCLUDED.team_id,
			is_active = EXCLUDED.is_active;
	`

	sqlSelectUserByID = `
		SELECT u.user_id, u.username, t.name, u.is_active
		FROM users u
		JOIN teams t ON t.id = u.team_id
		WHERE u.user_id = $1;
	`

	sqlUpdateUserIsActive = `
		UPDATE users u
		SET is_active = $2
		FROM teams t
		WHERE u.team_id = t.id
		  AND u.user_id = $1
		RETURNING u.user_id, u.username, t.name, u.is_active;
	`

	sqlListActiveUsersByTeamName = `
		SELECT u.user_id, u.username, t.name, u.is_active
		FROM users u
		JOIN teams t ON t.id = u.team_id
		WHERE t.name = $1
		  AND u.is_active = TRUE
		ORDER BY u.user_id;
	`
)

func (r *userRepository) UpsertMany(ctx context.Context, teamName string, users []domain.User) error {
	q := r.tx.Querier(ctx)

	var teamID int64
	if err := q.QueryRow(ctx, sqlSelectTeamIDByName, teamName).Scan(&teamID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("team not found")
		}
		return fmt.Errorf("get team_id for team %q: %w", teamName, err)
	}

	for _, u := range users {
		if _, err := q.Exec(ctx, sqlUpsertUser, u.ID, u.Username, teamID, u.IsActive); err != nil {
			return fmt.Errorf("upsert user %q for team %q: %w", u.ID, teamName, err)
		}
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, userID string) (domain.User, error) {
	q := r.tx.Querier(ctx)

	var (
		id       string
		username string
		teamName string
		isActive bool
	)

	if err := q.QueryRow(ctx, sqlSelectUserByID, userID).Scan(
		&id,
		&username,
		&teamName,
		&isActive,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.NotFound("user not found")
		}
		return domain.User{}, fmt.Errorf("get user %q: %w", userID, err)
	}

	return domain.User{
		ID:       id,
		Username: username,
		TeamName: teamName,
		IsActive: isActive,
	}, nil
}

func (r *userRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error) {
	q := r.tx.Querier(ctx)

	var (
		id       string
		username string
		teamName string
		active   bool
	)

	err := q.QueryRow(ctx, sqlUpdateUserIsActive, userID, isActive).Scan(
		&id,
		&username,
		&teamName,
		&active,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.NotFound("user not found")
		}
		return domain.User{}, fmt.Errorf("set is_active=%v for user %q: %w", isActive, userID, err)
	}

	return domain.User{
		ID:       id,
		Username: username,
		TeamName: teamName,
		IsActive: active,
	}, nil
}

func (r *userRepository) ListActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	q := r.tx.Querier(ctx)

	rows, err := q.Query(ctx, sqlListActiveUsersByTeamName, teamName)
	if err != nil {
		return nil, fmt.Errorf("list active users for team %q: %w", teamName, err)
	}
	defer rows.Close()

	users := make([]domain.User, 0)

	for rows.Next() {
		var (
			id       string
			username string
			tName    string
			active   bool
		)

		if err := rows.Scan(&id, &username, &tName, &active); err != nil {
			return nil, fmt.Errorf("scan active user for team %q: %w", teamName, err)
		}

		users = append(users, domain.User{
			ID:       id,
			Username: username,
			TeamName: tName,
			IsActive: active,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active users for team %q: %w", teamName, err)
	}

	return users, nil
}
