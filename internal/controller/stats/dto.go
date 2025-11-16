package stats

type ReviewerAssignmentStat struct {
	ReviewerID       string
	Username         string
	TeamName         string
	AssignmentsCount int64
}

type AssignmentsStatsOutput struct {
	Stats []ReviewerAssignmentStat
}
