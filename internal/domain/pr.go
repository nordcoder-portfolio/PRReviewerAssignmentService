package domain

import "time"

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            PRStatus
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          *time.Time
}

type PullRequestShort struct {
	ID       string
	Name     string
	AuthorID string
	Status   PRStatus
}
