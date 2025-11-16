package team

import (
	"avito_test/internal/domain"
	"context"
	"log/slog"
)

func (c *controller) CreateTeam(ctx context.Context, in CreateTeamInput) (Output, error) {
	log := c.logger.With(
		slog.String("op", "create_team"),
		slog.String("team_name", in.TeamName),
	)

	if err := validateCreateInput(in); err != nil {
		log.Debug("invalid create_team input", slog.Any("error", err))
		return Output{}, err
	}

	domainMembers := make([]domain.User, 0, len(in.Members))
	for _, m := range in.Members {
		domainMembers = append(domainMembers, domain.User{
			ID:       m.UserID,
			Username: m.Username,
			TeamName: in.TeamName,
			IsActive: m.IsActive,
		})
	}

	domainTeam := domain.Team{
		Name:    in.TeamName,
		Members: domainMembers,
	}

	created, err := c.uc.CreateTeam(ctx, domainTeam)
	if err != nil {
		return Output{}, err
	}

	return mapDomainTeamToOutput(created), nil
}

func (c *controller) GetTeam(ctx context.Context, teamName string) (Output, error) {
	log := c.logger.With(
		slog.String("op", "get_team"),
		slog.String("team_name", teamName),
	)

	if teamName == "" {
		log.Debug("empty team_name in get_team")
		return Output{}, domain.BadRequest("team_name must not be empty")
	}

	t, err := c.uc.GetTeam(ctx, teamName)
	if err != nil {
		return Output{}, err
	}

	return mapDomainTeamToOutput(t), nil
}

func (c *controller) DeactivateMembers(
	ctx context.Context,
	input DeactivateMembersInput,
) (DeactivateMembersOutput, error) {
	log := c.logger.With(
		slog.String("op", "deactivate_team_members"),
		slog.String("team", input.TeamName),
		slog.Int("user_count", len(input.UserIDs)),
	)

	teamDomain, updatedPRs, err := c.uc.DeactivateTeamMembersAndReassignOpenPRs(
		ctx,
		input.TeamName,
		input.UserIDs,
	)
	if err != nil {
		log.Error("usecase failed", slog.String("error", err.Error()))
		return DeactivateMembersOutput{}, err
	}

	outTeam := Output{
		TeamName: teamDomain.Name,
		Members:  make([]Member, len(teamDomain.Members)),
	}
	for i, m := range teamDomain.Members {
		outTeam.Members[i] = Member{
			UserID:   m.ID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	outPRs := make([]DeactivateMembersPROutput, len(updatedPRs))
	for i, pr := range updatedPRs {
		outPRs[i] = DeactivateMembersPROutput{
			ID:       pr.ID,
			Name:     pr.Name,
			AuthorID: pr.AuthorID,
			Status:   string(pr.Status),
		}
	}

	return DeactivateMembersOutput{
		Team:                outTeam,
		UpdatedPullRequests: outPRs,
	}, nil
}
