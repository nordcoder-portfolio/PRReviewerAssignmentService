package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateTeam_TeamExists(t *testing.T) {
	require.NotNil(t, httpClient, "client is nil")

	suffix := time.Now().UnixNano()
	teamName := fmt.Sprintf("backend-exists-e2e-%d", suffix)

	reqTeam := Team{
		TeamName: teamName,
		Members: []TeamMember{
			{UserID: simpleID1, Username: "Vasya", IsActive: true},
		},
	}

	var resp1 struct {
		Team Team `json:"team"`
	}

	status, err := httpClient.DoJSON(
		http.MethodPost,
		"/team/add",
		nil,
		reqTeam,
		http.StatusCreated,
		&resp1,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, status)

	var errResp ErrorResponse

	status, err = httpClient.DoJSON(
		http.MethodPost,
		"/team/add",
		nil,
		reqTeam,
		http.StatusBadRequest,
		&errResp,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, "TEAM_EXISTS", errResp.Error.Code)
}
