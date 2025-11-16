package http

import (
	"avito_test/internal/api"
	"avito_test/internal/controller/pr"
	"encoding/json"
	"net/http"
)

func (s *Server) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body api.PostPullRequestCreateJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeBadRequest(w, "invalid request body")
		return
	}

	out, err := s.prCtrl.Create(ctx, pr.CreateInput{
		ID:       body.PullRequestId,
		Name:     body.PullRequestName,
		AuthorID: body.AuthorId,
	})
	if err != nil {
		s.writeError(w, err)
		return
	}

	respPR := mapPRToAPI(out.PR)

	s.writeJSON(w, http.StatusCreated, map[string]any{
		"pr": respPR,
	})
}

func (s *Server) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body api.PostPullRequestMergeJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeBadRequest(w, "invalid request body")
		return
	}

	out, err := s.prCtrl.Merge(ctx, pr.MergeInput{
		ID: body.PullRequestId,
	})
	if err != nil {
		s.writeError(w, err)
		return
	}

	respPR := mapPRToAPI(out.PR)

	s.writeJSON(w, http.StatusOK, map[string]any{
		"pr": respPR,
	})
}

func (s *Server) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body api.PostPullRequestReassignJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeBadRequest(w, "invalid request body")
		return
	}

	out, err := s.prCtrl.Reassign(ctx, pr.ReassignInput{
		ID:        body.PullRequestId,
		OldUserID: body.OldUserId,
	})
	if err != nil {
		s.writeError(w, err)
		return
	}

	respPR := mapPRToAPI(out.PR)

	s.writeJSON(w, http.StatusOK, map[string]any{
		"pr":          respPR,
		"replaced_by": out.ReplacedBy,
	})
}

func mapPRToAPI(p pr.PR) api.PullRequest {
	return api.PullRequest{
		PullRequestId:     p.ID,
		PullRequestName:   p.Name,
		AuthorId:          p.AuthorID,
		Status:            api.PullRequestStatus(p.Status),
		AssignedReviewers: append([]string(nil), p.AssignedReviewers...),
		CreatedAt:         nil,
		MergedAt:          nil,
	}
}
