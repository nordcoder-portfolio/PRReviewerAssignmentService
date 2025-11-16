package stats

import (
	"avito_test/internal/domain"
	"context"
	"log/slog"
	"sort"
)

func (u *usecase) GetReviewerAssignmentsStats(ctx context.Context) ([]domain.ReviewerAssignmentStat, error) {
	stats, err := u.repo.ListReviewerAssignmentsStats(ctx)

	if err != nil {
		u.logger.Error("get reviewer stats failed", slog.String("error", err.Error()))
		return nil, err
	}

	result := make([]domain.ReviewerAssignmentStat, 0, len(stats))

	for user, count := range stats {
		result = append(result, domain.ReviewerAssignmentStat{
			ReviewerID:       user.ID,
			Username:         user.Username,
			TeamName:         user.TeamName,
			AssignmentsCount: count,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].AssignmentsCount < result[j].AssignmentsCount
	})

	return result, nil
}
