package team

import (
	"avito_test/internal/domain"
	helpers "avito_test/internal/usecase"
	pruc "avito_test/internal/usecase/pr"
	"context"
	"fmt"
	"log/slog"
)

func (uc *usecase) CreateTeam(ctx context.Context, team domain.Team) (domain.Team, error) {
	log := uc.logger.With(
		slog.String("op", "create_team"),
		slog.String("team_name", team.Name),
	)

	normalizeTeamMembers(&team)

	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.teams.Create(ctx, team.Name); err != nil {
			return helpers.ConvertAndLogError(log, "create team", err)
		}

		if err := uc.users.UpsertMany(ctx, team.Name, team.Members); err != nil {
			return helpers.ConvertAndLogError(log, "upsert team members", err)
		}

		return nil
	})
	if err != nil {
		return domain.Team{}, err
	}

	log.Info("team created",
		slog.Int("members_count", len(team.Members)),
	)

	return team, nil
}

func (uc *usecase) GetTeam(ctx context.Context, name string) (domain.Team, error) {
	log := uc.logger.With(
		slog.String("op", "get_team"),
		slog.String("team_name", name),
	)

	team, err := uc.teams.GetByName(ctx, name)
	if err != nil {
		return domain.Team{}, helpers.ConvertAndLogError(log, "get team by name", err)
	}

	return team, nil
}

func (uc *usecase) DeactivateTeamMembersAndReassignOpenPRs(
	ctx context.Context,
	teamName string,
	userIDs []string,
) (domain.Team, []domain.PullRequestShort, error) {
	log := uc.logger.With(
		slog.String("op", "deactivate_team_members"),
		slog.String("team_name", teamName),
		slog.Int("user_ids_count", len(userIDs)),
	)

	uniqueIDs, deactivatedSet := uniqueStringSet(userIDs)
	if len(uniqueIDs) == 0 {
		team, err := uc.teams.GetByName(ctx, teamName)
		if err != nil {
			return domain.Team{}, nil, helpers.ConvertAndLogError(log, "get team by name", err)
		}
		return team, nil, nil
	}

	var (
		resultTeam domain.Team
		updatedPRs []domain.PullRequestShort
	)

	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		team, err := uc.teams.GetByName(ctx, teamName)
		if err != nil {
			return helpers.ConvertAndLogError(log, "get team by name", err)
		}

		memberByID, candidateIDs := buildTeamState(team, deactivatedSet)

		for _, id := range uniqueIDs {
			if _, ok := memberByID[id]; !ok {
				return helpers.ConvertAndLogError(
					log,
					"validate deactivation users",
					domain.NotFound(fmt.Sprintf("user %q is not a member of team %q", id, teamName)),
				)
			}
		}

		prIDs, err := uc.collectOpenPRIDs(ctx, uniqueIDs)
		if err != nil {
			return helpers.ConvertAndLogError(log, "collect open PRs for reviewers", err)
		}

		updatedPRs = make([]domain.PullRequestShort, 0, len(prIDs))

		for prID := range prIDs {
			pr, errGetbyid := uc.prs.GetByID(ctx, prID)
			if errGetbyid != nil {
				return helpers.ConvertAndLogError(log, "get PR by id", errGetbyid)
			}
			if pr.Status != domain.PRStatusOpen {
				continue
			}

			changed := reassignReviewers(&pr, deactivatedSet, candidateIDs, uc.chooser)
			if !changed {
				continue
			}

			if errUpdate := uc.prs.Update(ctx, pr); errUpdate != nil {
				return helpers.ConvertAndLogError(log, "update PR", errUpdate)
			}

			updatedPRs = append(updatedPRs, domain.PullRequestShort{
				ID:       pr.ID,
				Name:     pr.Name,
				AuthorID: pr.AuthorID,
				Status:   pr.Status,
			})
		}

		for _, id := range uniqueIDs {
			if _, errSetActive := uc.users.SetIsActive(ctx, id, false); errSetActive != nil {
				return helpers.ConvertAndLogError(log, "deactivate user", errSetActive)
			}
		}

		teamAfter, err := uc.teams.GetByName(ctx, teamName)
		if err != nil {
			return helpers.ConvertAndLogError(log, "get team after deactivation", err)
		}

		resultTeam = teamAfter
		return nil
	})
	if err != nil {
		return domain.Team{}, nil, err
	}

	log.Info("team members deactivated and PRs reassigned",
		slog.Int("deactivated_users", len(uniqueIDs)),
		slog.Int("updated_prs", len(updatedPRs)),
	)

	return resultTeam, updatedPRs, nil
}

func normalizeTeamMembers(team *domain.Team) {
	for i := range team.Members {
		team.Members[i].TeamName = team.Name
	}
}

type stringSet map[string]struct{}

func uniqueStringSet(ids []string) ([]string, stringSet) {
	set := make(stringSet, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := set[id]; ok {
			continue
		}
		set[id] = struct{}{}
		out = append(out, id)
	}
	return out, set
}

func buildTeamState(team domain.Team, deactivated stringSet) (map[string]domain.User, []string) {
	memberByID := make(map[string]domain.User, len(team.Members))
	candidates := make([]string, 0, len(team.Members))

	for _, m := range team.Members {
		memberByID[m.ID] = m
		if !m.IsActive {
			continue
		}
		if _, deact := deactivated[m.ID]; deact {
			continue
		}
		candidates = append(candidates, m.ID)
	}

	return memberByID, candidates
}

func (uc *usecase) collectOpenPRIDs(
	ctx context.Context,
	reviewerIDs []string,
) (map[string]struct{}, error) {
	prIDs := make(map[string]struct{})

	for _, reviewerID := range reviewerIDs {
		prs, err := uc.prs.ListByReviewer(ctx, reviewerID)
		if err != nil {
			return nil, err
		}
		for _, pr := range prs {
			if pr.Status != domain.PRStatusOpen {
				continue
			}
			prIDs[pr.ID] = struct{}{}
		}
	}

	return prIDs, nil
}

func reassignReviewers(
	pr *domain.PullRequest,
	deactivated stringSet,
	globalCandidates []string,
	chooser reviewerChooser,
) bool {
	seen := make(stringSet, len(pr.AssignedReviewers))
	newAssigned := make([]string, 0, len(pr.AssignedReviewers))

	for _, id := range pr.AssignedReviewers {
		if _, deact := deactivated[id]; deact {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		newAssigned = append(newAssigned, id)
	}

	if len(newAssigned) > pruc.PrReviewersCount {
		newAssigned = newAssigned[:pruc.PrReviewersCount]
	}

	need := pruc.PrReviewersCount - len(newAssigned)
	if need > 0 && chooser != nil && len(globalCandidates) > 0 {
		forbidden := make(stringSet, len(seen)+1)
		forbidden[pr.AuthorID] = struct{}{}
		for id := range seen {
			forbidden[id] = struct{}{}
		}

		localCandidates := make([]string, 0, len(globalCandidates))
		for _, id := range globalCandidates {
			if _, bad := forbidden[id]; bad {
				continue
			}
			localCandidates = append(localCandidates, id)
		}

		chosen := chooser.Choice(localCandidates, need)
		for _, id := range chosen {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			newAssigned = append(newAssigned, id)
		}
	}

	changed := !equalStringSlices(pr.AssignedReviewers, newAssigned)
	pr.AssignedReviewers = newAssigned
	return changed
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
