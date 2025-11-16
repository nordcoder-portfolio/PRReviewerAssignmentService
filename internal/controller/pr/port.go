package pr

import (
	"avito_test/internal/domain"
	"context"
	"log/slog"
)

var _ Controller = (*controller)(nil)

type Controller interface {
	Create(ctx context.Context, in CreateInput) (CreateOutput, error)
	Merge(ctx context.Context, in MergeInput) (MergeOutput, error)
	Reassign(ctx context.Context, in ReassignInput) (ReassignOutput, error)
}

type usecase interface {
	Create(ctx context.Context, id, name, authorID string) (domain.PullRequest, error)
	Merge(ctx context.Context, id string) (domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (domain.PullRequest, string, error)
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
