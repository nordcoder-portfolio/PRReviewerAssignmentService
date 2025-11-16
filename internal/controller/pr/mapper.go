package pr

import "avito_test/internal/domain"

func mapDomainPR(pr domain.PullRequest) PR {
	return PR{
		ID:                pr.ID,
		Name:              pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: append([]string(nil), pr.AssignedReviewers...),
	}
}
