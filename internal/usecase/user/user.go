package user

import (
	"avito_test/internal/domain"
	helpers "avito_test/internal/usecase"
	"context"
	"log/slog"
)

func (uc *usecase) SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error) {
	log := uc.logger.With(
		slog.String("op", "set_is_active"),
		slog.String("user_id", userID),
		slog.Bool("is_active", isActive),
	)

	user, err := uc.users.SetIsActive(ctx, userID, isActive)
	if err != nil {
		return domain.User{}, helpers.ConvertAndLogError(log, "set user activity", err)
	}

	log.Info("user activity updated")

	return user, nil
}

func (uc *usecase) GetReviewPRs(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	log := uc.logger.With(
		slog.String("op", "get_review_prs"),
		slog.String("user_id", userID),
	)

	prs, err := uc.prs.ListByReviewer(ctx, userID)
	if err != nil {
		return nil, helpers.ConvertAndLogError(log, "list pull requests for reviewer", err)
	}

	return prs, nil
}
