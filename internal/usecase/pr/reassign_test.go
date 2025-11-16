package pr

import (
	"avito_test/internal/domain"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestUsecase_ReassignReviewer_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	const (
		prID      = "pr-1"
		oldUserID = "u2"
		teamName  = "backend"
	)

	uc, users, prs, chooser := newPRUsecase()

	existing := domain.PullRequest{
		ID:                prID,
		Name:              "Feature",
		AuthorID:          "author-1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u1", "u2"},
	}

	oldUser := domain.User{
		ID:       oldUserID,
		TeamName: teamName,
	}

	members := []domain.User{
		{ID: "u1", TeamName: teamName},
		{ID: "u2", TeamName: teamName},
		{ID: "u3", TeamName: teamName},
	}

	prs.
		On("GetByID", mock.Anything, prID).
		Return(existing, nil)

	users.
		On("GetByID", mock.Anything, oldUserID).
		Return(oldUser, nil)

	users.
		On("ListActiveByTeam", mock.Anything, teamName).
		Return(members, nil)

	chooser.result = []string{"u3"}

	expected := existing
	expected.AssignedReviewers = []string{"u1", "u3"}

	prs.
		On("Update", mock.Anything, mock.MatchedBy(func(pr domain.PullRequest) bool {
			if pr.ID != expected.ID || pr.Status != expected.Status {
				return false
			}
			return reflect.DeepEqual(pr.AssignedReviewers, expected.AssignedReviewers)
		})).
		Return(nil)

	got, replacedBy, err := uc.ReassignReviewer(ctx, prID, oldUserID)
	if err != nil {
		t.Fatalf("ReassignReviewer() got err: %v, want nil", err)
	}

	if replacedBy != "u3" {
		t.Fatalf("ReassignReviewer() replacedBy = %q, want %q", replacedBy, "u3")
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("ReassignReviewer() got = %#v, want %#v", got, expected)
	}

	if !chooser.called {
		t.Fatalf("ReassignReviewer() chooser should be called")
	}
	if chooser.limit != 1 {
		t.Fatalf("ReassignReviewer() chooser.limit = %d, want 1", chooser.limit)
	}
	if !reflect.DeepEqual(chooser.candidates, []string{"u3"}) {
		t.Fatalf("ReassignReviewer() candidates = %v, want [u3]", chooser.candidates)
	}

	prs.AssertExpectations(t)
	users.AssertExpectations(t)
}

func TestUsecase_ReassignReviewer_PRMerged(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	const (
		prID      = "pr-1"
		oldUserID = "u2"
	)

	uc, users, prs, chooser := newPRUsecase()

	existing := domain.PullRequest{
		ID:                prID,
		Name:              "Feature",
		AuthorID:          "author-1",
		Status:            domain.PRStatusMerged,
		AssignedReviewers: []string{"u1", "u2"},
	}

	prs.
		On("GetByID", mock.Anything, prID).
		Return(existing, nil)

	got, replacedBy, err := uc.ReassignReviewer(ctx, prID, oldUserID)

	appErr := domain.PRMerged("cannot reassign on merged PR")
	if !errors.Is(err, appErr) {
		t.Fatalf("ReassignReviewer() got err: %v, want %v", err, appErr)
	}

	assertZeroPR(t, got)
	if replacedBy != "" {
		t.Fatalf("ReassignReviewer() replacedBy = %q, want empty", replacedBy)
	}

	users.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
	users.AssertNotCalled(t, "ListActiveByTeam", mock.Anything, mock.Anything)
	prs.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	if chooser.called {
		t.Fatalf("ReassignReviewer() chooser should not be called for merged PR")
	}

	prs.AssertExpectations(t)
	users.AssertExpectations(t)
}

func TestUsecase_ReassignReviewer_NotAssigned(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	const (
		prID      = "pr-1"
		oldUserID = "u2"
	)

	uc, users, prs, chooser := newPRUsecase()

	existing := domain.PullRequest{
		ID:                prID,
		Name:              "Feature",
		AuthorID:          "author-1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u1", "u3"},
	}

	prs.
		On("GetByID", mock.Anything, prID).
		Return(existing, nil)

	got, replacedBy, err := uc.ReassignReviewer(ctx, prID, oldUserID)

	appErr := domain.NotAssigned("reviewer is not assigned to this PR")
	if !errors.Is(err, appErr) {
		t.Fatalf("ReassignReviewer() got err: %v, want %v", err, appErr)
	}

	assertZeroPR(t, got)
	if replacedBy != "" {
		t.Fatalf("ReassignReviewer() replacedBy = %q, want empty", replacedBy)
	}

	users.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
	users.AssertNotCalled(t, "ListActiveByTeam", mock.Anything, mock.Anything)
	prs.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	if chooser.called {
		t.Fatalf("ReassignReviewer() chooser should not be called when reviewer not assigned")
	}

	prs.AssertExpectations(t)
	users.AssertExpectations(t)
}

func TestUsecase_ReassignReviewer_NoCandidate(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	const (
		prID      = "pr-1"
		oldUserID = "u2"
		teamName  = "backend"
	)

	uc, users, prs, _ := newPRUsecase()

	existing := domain.PullRequest{
		ID:                prID,
		Name:              "Feature",
		AuthorID:          "author-1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u1", "u2"},
	}

	oldUser := domain.User{
		ID:       oldUserID,
		TeamName: teamName,
	}

	members := []domain.User{
		{ID: "u1", TeamName: teamName},
		{ID: "u2", TeamName: teamName},
	}

	prs.
		On("GetByID", mock.Anything, prID).
		Return(existing, nil)

	users.
		On("GetByID", mock.Anything, oldUserID).
		Return(oldUser, nil)

	users.
		On("ListActiveByTeam", mock.Anything, teamName).
		Return(members, nil)

	got, replacedBy, err := uc.ReassignReviewer(ctx, prID, oldUserID)

	appErr := domain.NoCandidate("no active replacement candidate in team")
	if !errors.Is(err, appErr) {
		t.Fatalf("ReassignReviewer() got err: %v, want %v", err, appErr)
	}

	assertZeroPR(t, got)
	if replacedBy != "" {
		t.Fatalf("ReassignReviewer() replacedBy = %q, want empty", replacedBy)
	}

	prs.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	prs.AssertExpectations(t)
	users.AssertExpectations(t)
}

func TestUsecase_ReassignReviewer_Update_AppError(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	const (
		prID      = "pr-1"
		oldUserID = "u2"
		teamName  = "backend"
	)

	uc, users, prs, chooser := newPRUsecase()

	existing := domain.PullRequest{
		ID:                prID,
		Name:              "Feature",
		AuthorID:          "author-1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u1", "u2"},
	}

	oldUser := domain.User{
		ID:       oldUserID,
		TeamName: teamName,
	}

	members := []domain.User{
		{ID: "u1", TeamName: teamName},
		{ID: "u2", TeamName: teamName},
		{ID: "u3", TeamName: teamName},
	}

	prs.
		On("GetByID", mock.Anything, prID).
		Return(existing, nil)

	users.
		On("GetByID", mock.Anything, oldUserID).
		Return(oldUser, nil)

	users.
		On("ListActiveByTeam", mock.Anything, teamName).
		Return(members, nil)

	chooser.result = []string{"u3"}

	appErr := domain.PRMerged("update failed")

	prs.
		On("Update", mock.Anything, mock.Anything).
		Return(appErr)

	got, replacedBy, err := uc.ReassignReviewer(ctx, prID, oldUserID)

	if !errors.Is(err, appErr) {
		t.Fatalf("ReassignReviewer() got err: %v, want %v", err, appErr)
	}

	assertZeroPR(t, got)
	if replacedBy != "" {
		t.Fatalf("ReassignReviewer() replacedBy = %q, want empty", replacedBy)
	}

	prs.AssertExpectations(t)
	users.AssertExpectations(t)
}
