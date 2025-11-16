package pr

import (
	"context"
	"log/slog"
)

func (c *controller) Create(ctx context.Context, in CreateInput) (CreateOutput, error) {
	log := c.logger.With(
		slog.String("op", "create_pr"),
		slog.String("pr_id", in.ID),
		slog.String("author_id", in.AuthorID),
	)

	if err := validateCreateInput(in); err != nil {
		log.Debug("invalid create_pr input", slog.Any("error", err))
		return CreateOutput{}, err
	}

	pr, err := c.uc.Create(ctx, in.ID, in.Name, in.AuthorID)
	if err != nil {
		return CreateOutput{}, err
	}

	return CreateOutput{
		PR: mapDomainPR(pr),
	}, nil
}

func (c *controller) Merge(ctx context.Context, in MergeInput) (MergeOutput, error) {
	log := c.logger.With(
		slog.String("op", "merge_pr"),
		slog.String("pr_id", in.ID),
	)

	if err := validateMergeInput(in); err != nil {
		log.Debug("invalid merge_pr input", slog.Any("error", err))
		return MergeOutput{}, err
	}

	pr, err := c.uc.Merge(ctx, in.ID)
	if err != nil {
		return MergeOutput{}, err
	}

	return MergeOutput{
		PR: mapDomainPR(pr),
	}, nil
}

func (c *controller) Reassign(ctx context.Context, in ReassignInput) (ReassignOutput, error) {
	log := c.logger.With(
		slog.String("op", "reassign_reviewer"),
		slog.String("pr_id", in.ID),
		slog.String("old_reviewer_id", in.OldUserID),
	)

	if err := validateReassignInput(in); err != nil {
		log.Debug("invalid reassign_reviewer input", slog.Any("error", err))
		return ReassignOutput{}, err
	}

	pr, replacedBy, err := c.uc.ReassignReviewer(ctx, in.ID, in.OldUserID)
	if err != nil {
		return ReassignOutput{}, err
	}

	return ReassignOutput{
		PR:         mapDomainPR(pr),
		ReplacedBy: replacedBy,
	}, nil
}
