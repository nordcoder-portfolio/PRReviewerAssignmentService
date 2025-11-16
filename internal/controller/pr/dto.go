package pr

type CreateInput struct {
	ID       string
	Name     string
	AuthorID string
}

type PR struct {
	ID                string
	Name              string
	AuthorID          string
	Status            string
	AssignedReviewers []string
}

type CreateOutput struct {
	PR PR
}

type MergeInput struct {
	ID string
}

type MergeOutput struct {
	PR PR
}

type ReassignInput struct {
	ID        string
	OldUserID string
}

type ReassignOutput struct {
	PR         PR
	ReplacedBy string
}
