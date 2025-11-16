package team

type Member struct {
	UserID   string
	Username string
	IsActive bool
}

type CreateTeamInput struct {
	TeamName string
	Members  []Member
}

type Output struct {
	TeamName string
	Members  []Member
}

type DeactivateMembersInput struct {
	TeamName string
	UserIDs  []string
}

type DeactivateMembersPROutput struct {
	ID       string
	Name     string
	AuthorID string
	Status   string
}

type DeactivateMembersOutput struct {
	Team                Output
	UpdatedPullRequests []DeactivateMembersPROutput
}
