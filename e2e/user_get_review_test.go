package e2e

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetUserReview_ReturnsAssignedPRs(t *testing.T) {
	require.NotNil(t, httpClient, "client is nil")

	suffix := time.Now().UnixNano()
	teamName := fmt.Sprintf("backend-review-e2e-%d", suffix)
	prID := fmt.Sprintf("pr-review-e2e-%d", suffix)

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

	reqPR := map[string]any{
		"pull_request_id":   prID,
		"pull_request_name": "E2E GetReview Test PR",
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

	require.NotEmpty(t, prResp.PR.AssignedReviewers, "reviewers must not be empty")

	reviewerID := prResp.PR.AssignedReviewers[0]

	q := url.Values{}
	q.Set("user_id", reviewerID)

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
	require.Equal(t, reviewerID, reviewsResp.UserID)

	found := false
	for _, pr := range reviewsResp.PullRequests {
		if pr.PullRequestID == prID {
			found = true
			require.Equal(t, "OPEN", pr.Status)
			require.Equal(t, authorID, pr.AuthorID)
			break
		}
	}
	require.True(t, found, "PR must be in user reviews")
}
