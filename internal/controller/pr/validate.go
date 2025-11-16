package pr

import "avito_test/internal/domain"

func validateCreateInput(in CreateInput) error {
	if in.ID == "" {
		return domain.BadRequest("pull_request_id must not be empty")
	}
	if in.Name == "" {
		return domain.BadRequest("pull_request_name must not be empty")
	}
	if in.AuthorID == "" {
		return domain.BadRequest("author_id must not be empty")
	}
	return nil
}

func validateMergeInput(in MergeInput) error {
	if in.ID == "" {
		return domain.BadRequest("pull_request_id must not be empty")
	}
	return nil
}

func validateReassignInput(in ReassignInput) error {
	if in.ID == "" {
		return domain.BadRequest("pull_request_id must not be empty")
	}
	if in.OldUserID == "" {
		return domain.BadRequest("old_user_id must not be empty")
	}
	return nil
}
