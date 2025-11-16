package http

import (
	"avito_test/internal/api"
	"avito_test/internal/controller/user"
	"encoding/json"
	"net/http"
)

func (s *Server) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body api.PostUsersSetIsActiveJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeBadRequest(w, "invalid request body")
		return
	}

	updated, err := s.userCtrl.SetIsActive(ctx, user.SetIsActiveInput{
		UserID:   body.UserId,
		IsActive: body.IsActive,
	})
	if err != nil {
		s.writeError(w, err)
		return
	}

	respUser := api.User{
		UserId:   updated.UserID,
		Username: updated.Username,
		TeamName: updated.TeamName,
		IsActive: updated.IsActive,
	}

	s.writeJSON(w, http.StatusOK, map[string]any{
		"user": respUser,
	})
}

func (s *Server) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params api.GetUsersGetReviewParams) {
	ctx := r.Context()

	prs, err := s.userCtrl.GetReviewPRs(ctx, user.GetReviewInput{UserID: params.UserId})
	if err != nil {
		s.writeError(w, err)
		return
	}

	resp := struct {
		UserID       string                 `json:"user_id"`
		PullRequests []api.PullRequestShort `json:"pull_requests"`
	}{
		UserID:       prs.UserID,
		PullRequests: make([]api.PullRequestShort, 0, len(prs.PullRequests)),
	}

	for _, p := range prs.PullRequests {
		resp.PullRequests = append(resp.PullRequests, api.PullRequestShort{
			PullRequestId:   p.ID,
			PullRequestName: p.Name,
			AuthorId:        p.AuthorID,
			Status:          api.PullRequestShortStatus(p.Status),
		})
	}

	s.writeJSON(w, http.StatusOK, resp)
}
