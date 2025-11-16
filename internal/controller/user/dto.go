package user

type SetIsActiveInput struct {
	UserID   string
	IsActive bool
}

type SetIsActiveOutput struct {
	UserID   string
	Username string
	TeamName string
	IsActive bool
}

type GetReviewInput struct {
	UserID string
}

type ReviewPROutput struct {
	ID       string
	Name     string
	AuthorID string
	Status   string
}

type GetReviewOutput struct {
	UserID       string
	PullRequests []ReviewPROutput
}
