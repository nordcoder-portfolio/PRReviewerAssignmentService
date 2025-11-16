package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createTeam(t *testing.T, teamName string) Team {
	t.Helper()

	reqBody := Team{
		TeamName: teamName,
		Members: []TeamMember{
			{
				UserID:   simpleID1,
				Username: "Vasya",
				IsActive: true,
			},
			{
				UserID:   simpleID2,
				Username: "Petya",
				IsActive: true,
			},
			{
				UserID:   "u3",
				Username: "Charlie",
				IsActive: true,
			},
		},
	}

	var resp struct {
		Team Team `json:"team"`
	}

	status, err := httpClient.DoJSON(
		http.MethodPost,
		"/team/add",
		nil,
		reqBody,
		http.StatusCreated,
		&resp,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, status)

	return resp.Team
}

func TestCreatePullRequest_AssignsReviewersFromAuthorsTeam(t *testing.T) {
	require.NotNil(t, httpClient, "httpClient is not initialized in TestMain")

	suffix := time.Now().UnixNano()
	teamName := fmt.Sprintf("backend-pr-e2e-%d", suffix)
	prID := fmt.Sprintf("pr-e2e-%d", suffix)

	team := createTeam(t, teamName)
	require.Equal(t, teamName, team.TeamName)

	authorID := team.Members[0].UserID
	candidate1 := team.Members[1].UserID
	candidate2 := team.Members[2].UserID

	reqBody := map[string]any{
		"pull_request_id":   prID,
		"pull_request_name": "E2E Test PR",
		"author_id":         authorID,
	}

	var createResp struct {
		PR PullRequest `json:"pr"`
	}

	status, err := httpClient.DoJSON(
		http.MethodPost,
		"/pullRequest/create",
		nil,
		reqBody,
		http.StatusCreated,
		&createResp,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, status)

	pr := createResp.PR

	require.Equal(t, prID, pr.PullRequestID)
	require.Equal(t, "E2E Test PR", pr.PullRequestName)
	require.Equal(t, authorID, pr.AuthorID)
	require.Equal(t, "OPEN", pr.Status)

	require.NotEmpty(t, pr.AssignedReviewers, "must at least one reviewer")
	require.LessOrEqual(t, len(pr.AssignedReviewers), 2, "must maximum 2 reviewers")

	allowed := map[string]bool{
		candidate1: true,
		candidate2: true,
	}

	for _, rid := range pr.AssignedReviewers {
		require.NotEqual(t, authorID, rid, "author must not be a reviewer")
		require.Truef(t, allowed[rid], "reviewer must be from author's team")
	}
}
