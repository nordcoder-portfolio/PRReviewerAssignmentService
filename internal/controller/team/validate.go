package team

import (
	"avito_test/internal/domain"
	"fmt"
)

func validateCreateInput(in CreateTeamInput) error {
	if in.TeamName == "" {
		return domain.BadRequest("team_name must not be empty")
	}
	if len(in.Members) == 0 {
		return domain.BadRequest("team must contain at least one member")
	}

	seen := make(map[string]struct{}, len(in.Members))
	for i, m := range in.Members {
		if m.UserID == "" {
			return domain.BadRequest(fmt.Sprintf("members[%d].user_id must not be empty", i))
		}
		if m.Username == "" {
			return domain.BadRequest(fmt.Sprintf("members[%d].username must not be empty", i))
		}
		if _, ok := seen[m.UserID]; ok {
			return domain.BadRequest(fmt.Sprintf("duplicate user_id in members: %q", m.UserID))
		}
		seen[m.UserID] = struct{}{}
	}

	return nil
}
