package http

import (
	"avito_test/internal/api"
	"avito_test/internal/controller/team"
	"encoding/json"
	"net/http"
)

func (s *Server) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body api.PostTeamAddJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeBadRequest(w, "invalid request body")
		return
	}

	members := make([]team.Member, 0, len(body.Members))
	for _, m := range body.Members {
		members = append(members, team.Member{
			UserID:   m.UserId,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	out, err := s.teamCtrl.CreateTeam(ctx, team.CreateTeamInput{
		TeamName: body.TeamName,
		Members:  members,
	})
	if err != nil {
		s.writeError(w, err)
		return
	}

	respTeam := api.Team{
		TeamName: out.TeamName,
		Members:  make([]api.TeamMember, 0, len(out.Members)),
	}
	for _, u := range out.Members {
		respTeam.Members = append(respTeam.Members, api.TeamMember{
			UserId:   u.UserID,
			Username: u.Username,
			IsActive: u.IsActive,
		})
	}

	s.writeJSON(w, http.StatusCreated, map[string]any{
		"team": respTeam,
	})
}

func (s *Server) GetTeamGet(w http.ResponseWriter, r *http.Request, params api.GetTeamGetParams) {
	ctx := r.Context()

	out, err := s.teamCtrl.GetTeam(ctx, params.TeamName)
	if err != nil {
		s.writeError(w, err)
		return
	}

	respTeam := api.Team{
		TeamName: out.TeamName,
		Members:  make([]api.TeamMember, 0, len(out.Members)),
	}
	for _, u := range out.Members {
		respTeam.Members = append(respTeam.Members, api.TeamMember{
			UserId:   u.UserID,
			Username: u.Username,
			IsActive: u.IsActive,
		})
	}

	s.writeJSON(w, http.StatusOK, respTeam)
}

func (s *Server) PostTeamDeactivateMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body api.PostTeamDeactivateMembersJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeBadRequest(w, "invalid request body")
		return
	}

	if body.TeamName == "" {
		s.writeBadRequest(w, "invalid request body")
		return
	}

	input := team.DeactivateMembersInput{
		TeamName: body.TeamName,
		UserIDs:  body.UserIds,
	}

	out, err := s.teamCtrl.DeactivateMembers(ctx, input)
	if err != nil {
		s.writeError(w, err)
		return
	}

	respTeam := api.Team{
		TeamName: out.Team.TeamName,
		Members:  make([]api.TeamMember, len(out.Team.Members)),
	}
	for i, m := range out.Team.Members {
		respTeam.Members[i] = api.TeamMember{
			UserId:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	var respPRs []api.PullRequestShort
	if len(out.UpdatedPullRequests) > 0 {
		respPRs = make([]api.PullRequestShort, len(out.UpdatedPullRequests))
		for i, pr := range out.UpdatedPullRequests {
			respPRs[i] = api.PullRequestShort{
				PullRequestId:   pr.ID,
				PullRequestName: pr.Name,
				AuthorId:        pr.AuthorID,
				Status:          api.PullRequestShortStatus(pr.Status),
			}
		}
	}

	resp := struct {
		Team                api.Team               `json:"team"`
		UpdatedPullRequests []api.PullRequestShort `json:"updated_pull_requests"`
	}{
		Team:                respTeam,
		UpdatedPullRequests: respPRs,
	}

	s.writeJSON(w, http.StatusOK, resp)
}
