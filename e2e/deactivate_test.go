package e2e

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTeamDeactivateMembers_ReassignOpenPRs(t *testing.T) {
	require.NotNil(t, httpClient, "client is nil")

	suffix := time.Now().UnixNano()
	teamName := fmt.Sprintf("backend-deact-e2e-%d", suffix)
	targetReviewer := simpleID2

	reqTeam := Team{
		TeamName: teamName,
		Members: []TeamMember{
			{UserID: simpleID1, Username: "Vasya", IsActive: true},
			{UserID: simpleID2, Username: "Petya", IsActive: true},
			{UserID: "u3", Username: "Charlie", IsActive: true},
		},
	}

	var teamResp struct {
		Team Team `json:"team"`
	}

	status, err := httpClient.DoJSON(
		http.MethodPost,
		"/team/add",
		nil,
		reqTeam,
		http.StatusCreated,
		&teamResp,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, status)

	authorID := simpleID1

	const maxPRs = 8
	affectedPRs := make(map[string]bool)

	for i := 0; i < maxPRs && len(affectedPRs) == 0; i++ {
		prID := fmt.Sprintf("pr-deact-e2e-%d-%d", suffix, i)

		reqPR := map[string]any{
			"pull_request_id":   prID,
			"pull_request_name": "E2E Deactivate Test PR",
			"author_id":         authorID,
		}

		var prResp struct {
			PR PullRequest `json:"pr"`
		}

		status, err = httpClient.DoJSON(
			http.MethodPost,
			"/pullRequest/create",
			nil,
			reqPR,
			http.StatusCreated,
			&prResp,
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, status)

		for _, rid := range prResp.PR.AssignedReviewers {
			if rid == targetReviewer {
				affectedPRs[prResp.PR.PullRequestID] = true
				break
			}
		}
	}

	if len(affectedPRs) == 0 {
		t.Skip("no PR got target reviewer; cannot test deactivateMembers")
	}

	reqDeactivate := map[string]any{
		"team_name": teamName,
		"user_ids":  []string{targetReviewer},
	}

	var deactResp struct {
		Team                Team               `json:"team"`
		UpdatedPullRequests []PullRequestShort `json:"updated_pull_requests"`
	}

	status, err = httpClient.DoJSON(
		http.MethodPost,
		"/team/deactivateMembers",
		nil,
		reqDeactivate,
		http.StatusOK,
		&deactResp,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)

	var foundDeactivated bool
	activeUsers := make([]string, 0)

	for _, m := range deactResp.Team.Members {
		if m.UserID == targetReviewer {
			foundDeactivated = true
			require.False(t, m.IsActive, "deactivated user must be inactive")
		} else if m.IsActive {
			activeUsers = append(activeUsers, m.UserID)
		}
	}
	require.True(t, foundDeactivated, "deactivated user is not in team")

	updatedSet := make(map[string]bool)
	for _, pr := range deactResp.UpdatedPullRequests {
		updatedSet[pr.PullRequestID] = true
	}

	for prID := range affectedPRs {
		require.Truef(t, updatedSet[prID], "PR %s must be in updated_pull_requests", prID)
	}

	{
		q := url.Values{}
		q.Set("user_id", targetReviewer)

		var reviewsResp struct {
			UserID       string             `json:"user_id"`
			PullRequests []PullRequestShort `json:"pull_requests"`
		}

		status, err = httpClient.DoJSON(
			http.MethodGet,
			"/users/getReview",
			q,
			nil,
			http.StatusOK,
			&reviewsResp,
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		for _, pr := range reviewsResp.PullRequests {
			require.NotEqualf(t, "OPEN", pr.Status, "deactivated user still has OPEN PR %s", pr.PullRequestID)
		}
	}

	openByActive := make(map[string]bool)

	for _, uid := range activeUsers {
		q := url.Values{}
		q.Set("user_id", uid)

		var rResp struct {
			UserID       string             `json:"user_id"`
			PullRequests []PullRequestShort `json:"pull_requests"`
		}

		status, err = httpClient.DoJSON(
			http.MethodGet,
			"/users/getReview",
			q,
			nil,
			http.StatusOK,
			&rResp,
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		for _, pr := range rResp.PullRequests {
			if pr.Status == "OPEN" {
				openByActive[pr.PullRequestID] = true
			}
		}
	}

	for prID := range affectedPRs {
		require.Truef(t, openByActive[prID], "PR %s must have an active reviewer after deactivation", prID)
	}
}
