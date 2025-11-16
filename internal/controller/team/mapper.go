package team

import "avito_test/internal/domain"

func mapDomainTeamToOutput(t domain.Team) Output {
	members := make([]Member, 0, len(t.Members))
	for _, u := range t.Members {
		members = append(members, Member{
			UserID:   u.ID,
			Username: u.Username,
			IsActive: u.IsActive,
		})
	}

	return Output{
		TeamName: t.Name,
		Members:  members,
	}
}
