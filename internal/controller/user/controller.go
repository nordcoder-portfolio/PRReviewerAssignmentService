package user

import (
	"context"
	"log/slog"
)

func (c *controller) SetIsActive(ctx context.Context, in SetIsActiveInput) (SetIsActiveOutput, error) {
	log := c.logger.With(
		slog.String("op", "set_is_active"),
		slog.String("user_id", in.UserID),
		slog.Bool("is_active", in.IsActive),
	)

	if err := validateSetActiveInput(in); err != nil {
		log.Debug("invalid set_is_active input", slog.Any("error", err))
		return SetIsActiveOutput{}, err
	}

	user, err := c.uc.SetIsActive(ctx, in.UserID, in.IsActive)
	if err != nil {
		return SetIsActiveOutput{}, err
	}

	log.Info("user activity updated via controller")

	return SetIsActiveOutput{
		UserID:   user.ID,
		Username: user.Username,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}, nil
}

func (c *controller) GetReviewPRs(ctx context.Context, in GetReviewInput) (GetReviewOutput, error) {
	log := c.logger.With(
		slog.String("op", "get_review_prs"),
		slog.String("user_id", in.UserID),
	)

	if err := validateGetReviewInput(in); err != nil {
		log.Debug("invalid get_review_prs input", slog.Any("error", err))
		return GetReviewOutput{}, err
	}

	prs, err := c.uc.GetReviewPRs(ctx, in.UserID)
	if err != nil {
		return GetReviewOutput{}, err
	}

	out := GetReviewOutput{
		UserID:       in.UserID,
		PullRequests: make([]ReviewPROutput, 0, len(prs)),
	}

	for _, pr := range prs {
		out.PullRequests = append(out.PullRequests, ReviewPROutput{
			ID:       pr.ID,
			Name:     pr.Name,
			AuthorID: pr.AuthorID,
			Status:   string(pr.Status),
		})
	}
	return out, nil
}
