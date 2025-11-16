package http

import (
	"avito_test/internal/api"
	"avito_test/internal/controller/stats"
	"net/http"
)

func (s *Server) GetReviewerAssignmentsStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	out, err := s.statsCtrl.GetReviewerAssignmentsStats(ctx)
	if err != nil {
		s.writeError(w, err)
		return
	}

	resp := make([]api.ReviewerAssignmentStat, 0, len(out.Stats))
	for _, st := range out.Stats {
		resp = append(resp, mapReviewerAssignmentStatToAPI(st))
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func mapReviewerAssignmentStatToAPI(s stats.ReviewerAssignmentStat) api.ReviewerAssignmentStat {
	return api.ReviewerAssignmentStat{
		ReviewerId:       s.ReviewerID,
		Username:         &s.Username,
		TeamName:         &s.TeamName,
		AssignmentsCount: s.AssignmentsCount,
	}
}
