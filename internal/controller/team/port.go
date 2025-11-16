package team

import (
	"avito_test/internal/domain"
	"context"
	"log/slog"
)

var _ Controller = (*controller)(nil)

type Controller interface {
	CreateTeam(ctx context.Context, in CreateTeamInput) (Output, error)
	GetTeam(ctx context.Context, teamName string) (Output, error)
	DeactivateMembers(ctx context.Context,
		input DeactivateMembersInput) (DeactivateMembersOutput, error)
}

type usecase interface {
	CreateTeam(ctx context.Context, team domain.Team) (domain.Team, error)
	GetTeam(ctx context.Context, name string) (domain.Team, error)
	DeactivateTeamMembersAndReassignOpenPRs(
		ctx context.Context,
		teamName string,
		userIDs []string,
	) (domain.Team, []domain.PullRequestShort, error)
}

type controller struct {
	uc     usecase
	logger *slog.Logger
}

func New(uc usecase, logger *slog.Logger) *controller {
	return &controller{
		uc:     uc,
		logger: logger,
	}
}
