package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMergeIsIdempotent(t *testing.T) {
	require.NotNil(t, httpClient)

	suffix := time.Now().UnixNano()
	teamName := fmt.Sprintf("backend-merge-e2e-%d", suffix)
	prID := fmt.Sprintf("pr-merge-e2e-%d", suffix)

	team := createTeam(t, teamName)
	require.Equal(t, teamName, team.TeamName)
	authorID := team.Members[0].UserID

	reqPR := map[string]any{
		"pull_request_id":   prID,
		"pull_request_name": "E2E Merge Test PR",
		"author_id":         authorID,
	}

	var createResp struct {
		PR PullRequest `json:"pr"`
	}

	status, err := httpClient.DoJSON(
		http.MethodPost,
		"/pullRequest/create",
		nil,
		reqPR,
		http.StatusCreated,
		&createResp,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, status)

	orig := createResp.PR
	origReviewers := append([]string(nil), orig.AssignedReviewers...)

	reqMerge := map[string]any{
		"pull_request_id": prID,
	}

	var mergeResp1 struct {
		PR PullRequest `json:"pr"`
	}

	status, err = httpClient.DoJSON(
		http.MethodPost,
		"/pullRequest/merge",
		nil,
		reqMerge,
		http.StatusOK,
		&mergeResp1,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "MERGED", mergeResp1.PR.Status)
	require.Equal(t, prID, mergeResp1.PR.PullRequestID)
	require.Equal(t, authorID, mergeResp1.PR.AuthorID)
	require.ElementsMatch(t, origReviewers, mergeResp1.PR.AssignedReviewers)

	var mergeResp2 struct {
		PR PullRequest `json:"pr"`
	}

	status, err = httpClient.DoJSON(
		http.MethodPost,
		"/pullRequest/merge",
		nil,
		reqMerge,
		http.StatusOK,
		&mergeResp2,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "MERGED", mergeResp2.PR.Status)
	require.Equal(t, prID, mergeResp2.PR.PullRequestID)
	require.Equal(t, authorID, mergeResp2.PR.AuthorID)
	require.ElementsMatch(t, origReviewers, mergeResp2.PR.AssignedReviewers)
}
