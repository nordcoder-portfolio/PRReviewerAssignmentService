package stats

import (
	"context"
	"log/slog"
)

func (c *controller) GetReviewerAssignmentsStats(ctx context.Context) (AssignmentsStatsOutput, error) {
	stats, err := c.uc.GetReviewerAssignmentsStats(ctx)
	if err != nil {
		c.logger.Error("get reviewer stats failed", slog.String("error", err.Error()))
		return AssignmentsStatsOutput{}, err
	}

	out := AssignmentsStatsOutput{
		Stats: make([]ReviewerAssignmentStat, 0, len(stats)),
	}

	for _, s := range stats {
		out.Stats = append(out.Stats, ReviewerAssignmentStat{
			ReviewerID:       s.ReviewerID,
			Username:         s.Username,
			TeamName:         s.TeamName,
			AssignmentsCount: s.AssignmentsCount,
		})
	}

	return out, nil
}
