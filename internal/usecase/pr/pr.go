//go:build !mockery

package pr

import (
	"avito_test/internal/domain"
	helpers "avito_test/internal/usecase"
	"context"
	"log/slog"
)

const (
	PrReviewersCount = 2 // todo move to config
)

func (u *usecase) Create(ctx context.Context, id, name, authorID string) (domain.PullRequest, error) {
	log := u.logger.With(
		slog.String("op", "create_pr"),
		slog.String("pr_id", id),
		slog.String("author_id", authorID),
	)

	var result domain.PullRequest

	err := u.tx.WithinTx(ctx, func(ctx context.Context) error {
		author, err := u.getUserByID(ctx, log, authorID, "get author")
		if err != nil {
			return err
		}

		members, err := u.listActiveTeamMembers(ctx, log, author.TeamName,
			"list active team members")
		if err != nil {
			return err
		}

		candidates := buildReviewerCandidates(members, author.ID)
		reviewers := u.chooser.Choice(candidates, PrReviewersCount)

		pr := domain.PullRequest{
			ID:                id,
			Name:              name,
			AuthorID:          authorID,
			Status:            domain.PRStatusOpen,
			AssignedReviewers: reviewers,
		}

		if err := u.createPR(ctx, log, pr); err != nil {
			return err
		}

		result = pr

		log.Info("pull request created",
			slog.Int("reviewers_count", len(reviewers)),
		)

		return nil
	})

	return result, err
}

func (u *usecase) Merge(ctx context.Context, id string) (domain.PullRequest, error) {
	log := u.logger.With(
		slog.String("op", "merge_pr"),
		slog.String("pr_id", id),
	)

	var result domain.PullRequest

	err := u.tx.WithinTx(ctx, func(ctx context.Context) error {
		pr, err := u.getPRByID(ctx, log, id, "get pull request for merge")
		if err != nil {
			return err
		}

		if pr.Status == domain.PRStatusMerged {
			result = pr
			return nil
		}

		pr.Status = domain.PRStatusMerged

		if err := u.updatePR(ctx, log, pr, "update pull request status to MERGED"); err != nil {
			return err
		}

		result = pr
		log.Info("pull request merged")

		return nil
	})

	return result, err
}

func (u *usecase) ReassignReviewer(ctx context.Context, prID, oldUserID string) (domain.PullRequest, string, error) {
	log := u.logger.With(
		slog.String("op", "reassign_reviewer"),
		slog.String("pr_id", prID),
		slog.String("old_reviewer_id", oldUserID),
	)

	var (
		result     domain.PullRequest
		replacedBy string
	)

	err := u.tx.WithinTx(ctx, func(ctx context.Context) error {
		pr, err := u.getPRByID(ctx, log, prID, "get pull request for reassign")
		if err != nil {
			return err
		}

		if pr.Status == domain.PRStatusMerged {
			return domain.PRMerged("cannot reassign on merged PR")
		}

		if !isReviewerAssigned(pr, oldUserID) {
			return domain.NotAssigned("reviewer is not assigned to this PR")
		}

		oldUser, err := u.getUserByID(ctx, log, oldUserID, "get old reviewer user")
		if err != nil {
			return err
		}

		members, err := u.listActiveTeamMembers(ctx, log, oldUser.TeamName,
			"list active users by team")
		if err != nil {
			return err
		}

		candidates := buildReassignCandidates(pr, members, oldUserID)
		newID, err := u.pickReplacementReviewer(candidates)
		if err != nil {
			return err
		}

		pr.AssignedReviewers = replaceReviewer(pr.AssignedReviewers, oldUserID, newID)

		if err := u.updatePR(ctx, log, pr, "update pull request with new reviewer"); err != nil {
			return err
		}

		result = pr
		replacedBy = newID

		log.Info("reviewer reassigned successfully",
			slog.String("new_reviewer_id", newID),
		)

		return nil
	})

	return result, replacedBy, err
}

func (u *usecase) getUserByID(
	ctx context.Context,
	log *slog.Logger,
	userID string,
	action string,
) (domain.User, error) {
	user, err := u.users.GetByID(ctx, userID)
	if err != nil {
		return domain.User{}, helpers.ConvertAndLogError(log, action, err)
	}

	return user, nil
}

func (u *usecase) listActiveTeamMembers(
	ctx context.Context,
	log *slog.Logger,
	teamName string,
	action string,
) ([]domain.User, error) {
	members, err := u.users.ListActiveByTeam(ctx, teamName)
	if err != nil {
		log = log.With(slog.String("team_name", teamName))
		return nil, helpers.ConvertAndLogError(log, action, err)
	}

	return members, nil
}

func buildReviewerCandidates(members []domain.User, authorID string) []string {
	candidates := make([]string, 0, len(members))

	for _, m := range members {
		if m.ID == authorID {
			continue
		}
		candidates = append(candidates, m.ID)
	}

	return candidates
}

func (u *usecase) createPR(
	ctx context.Context,
	log *slog.Logger,
	pr domain.PullRequest,
) error {
	if err := u.prs.Create(ctx, pr); err != nil {
		return helpers.ConvertAndLogError(log, "create pull request", err)
	}

	return nil
}

func (u *usecase) getPRByID(
	ctx context.Context,
	log *slog.Logger,
	prID string,
	action string,
) (domain.PullRequest, error) {
	pr, err := u.prs.GetByID(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, helpers.ConvertAndLogError(log, action, err)
	}

	return pr, nil
}

func (u *usecase) updatePR(
	ctx context.Context,
	log *slog.Logger,
	pr domain.PullRequest,
	action string,
) error {
	if err := u.prs.Update(ctx, pr); err != nil {
		return helpers.ConvertAndLogError(log, action, err)
	}

	return nil
}

func isReviewerAssigned(pr domain.PullRequest, userID string) bool {
	for _, id := range pr.AssignedReviewers {
		if id == userID {
			return true
		}
	}
	return false
}

func buildReassignCandidates(
	pr domain.PullRequest,
	members []domain.User,
	oldUserID string,
) []string {
	exclude := make(map[string]struct{}, len(pr.AssignedReviewers)+1)

	exclude[pr.AuthorID] = struct{}{}
	exclude[oldUserID] = struct{}{}
	for _, id := range pr.AssignedReviewers {
		exclude[id] = struct{}{}
	}

	candidates := make([]string, 0, len(members))
	for _, m := range members {
		if _, skip := exclude[m.ID]; skip {
			continue
		}
		candidates = append(candidates, m.ID)
	}

	return candidates
}

func (u *usecase) pickReplacementReviewer(candidates []string) (string, error) {
	chosen := u.chooser.Choice(candidates, 1)
	if len(chosen) == 0 {
		return "", domain.NoCandidate("no active replacement candidate in team")
	}
	return chosen[0], nil
}

func replaceReviewer(reviewers []string, oldID, newID string) []string {
	for i, id := range reviewers {
		if id == oldID {
			reviewers[i] = newID
			break
		}
	}
	return reviewers
}
