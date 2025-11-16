package user

import "avito_test/internal/domain"

func validateSetActiveInput(in SetIsActiveInput) error {
	if in.UserID == "" {
		return domain.BadRequest("user_id must not be empty")
	}
	return nil
}

func validateGetReviewInput(in GetReviewInput) error {
	if in.UserID == "" {
		return domain.BadRequest("user_id must not be empty")
	}
	return nil
}
