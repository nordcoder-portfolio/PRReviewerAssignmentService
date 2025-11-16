package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReassignReviewer_Success(t *testing.T) {
	require.NotNil(t, httpClient, "client is nil")

	suffix := time.Now().UnixNano()
	teamName := fmt.Sprintf("backend-reassign-e2e-%d", suffix)
	prID := fmt.Sprintf("pr-reassign-e2e-%d", suffix)

	reqTeam := Team{
		TeamName: teamName,
		Members: []TeamMember{
			{UserID: simpleID1, Username: "Alice", IsActive: true},
			{UserID: simpleID2, Username: "Bob", IsActive: true},
			{UserID: "u3", Username: "Charlie", IsActive: true},
			{UserID: "u4", Username: "Dave", IsActive: true},
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
	oldReviewerID := simpleID2

	reqPR := map[string]any{
		"pull_request_id":   prID,
		"pull_request_name": "E2E Reassign Test PR",
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

	hasOld := false
	for _, rid := range prResp.PR.AssignedReviewers {
		if rid == oldReviewerID {
			hasOld = true
			break
		}
	}
	if !hasOld {
		t.Skip("old reviewer is not assigned; cannot test reassign")
	}

	reqReassign := map[string]any{
		"pull_request_id": prID,
		"old_user_id":     oldReviewerID,
	}

	var reassignResp struct {
		PR         PullRequest `json:"pr"`
		ReplacedBy string      `json:"replaced_by"`
	}

	status, err = httpClient.DoJSON(
		http.MethodPost,
		"/pullRequest/reassign",
		nil,
		reqReassign,
		http.StatusOK,
		&reassignResp,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)

	newPR := reassignResp.PR
	newReviewer := reassignResp.ReplacedBy

	require.Equal(t, "OPEN", newPR.Status)
	require.NotEqual(t, oldReviewerID, newReviewer, "new reviewer must be different from old")
	require.NotEqual(t, authorID, newReviewer, "author must not be reviewer")

	hasOld = false
	hasNew := false
	for _, rid := range newPR.AssignedReviewers {
		if rid == oldReviewerID {
			hasOld = true
		}
		if rid == newReviewer {
			hasNew = true
		}
	}
	require.False(t, hasOld, "old reviewer must be removed")
	require.True(t, hasNew, "new reviewer must be assigned")
}
