package pr

import (
	"avito_test/internal/domain"
	"avito_test/internal/repo/postgres"
	"avito_test/mocks"
	"context"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type txMock struct {
	err error
}

func (m *txMock) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	if m.err != nil {
		return m.err
	}
	return fn(ctx)
}

func (m *txMock) Querier(ctx context.Context) postgres.Querier {
	return nil
}

type chooserMock struct {
	candidates []string
	limit      int
	result     []string
	called     bool
}

func (c *chooserMock) Choice(candidates []string, limit int) []string {
	c.called = true
	c.limit = limit
	c.candidates = append([]string(nil), candidates...)
	return append([]string(nil), c.result...)
}

func assertZeroPR(t *testing.T, got domain.PullRequest) {
	t.Helper()
	if !reflect.DeepEqual(got, domain.PullRequest{}) {
		t.Fatalf("want empty pr, got %#v", got)
	}
}

func newPRUsecase() (*usecase, *mocks.UserRepository, *mocks.PRRepository, *chooserMock) {
	users := &mocks.UserRepository{}
	prs := &mocks.PRRepository{}
	tx := &txMock{}
	chooser := &chooserMock{}

	uc := &usecase{
		users:   users,
		prs:     prs,
		tx:      tx,
		chooser: chooser,
		logger:  newTestLogger(),
	}

	return uc, users, prs, chooser
}

func TestUsecase_Create_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	const (
		prID     = "pr-1"
		prName   = "Add feature"
		authorID = "user-1"
		teamName = "backend"
	)

	uc, users, prs, chooser := newPRUsecase()

	author := domain.User{
		ID:       authorID,
		TeamName: teamName,
	}
	members := []domain.User{
		{ID: authorID, TeamName: teamName},
		{ID: "u2", TeamName: teamName},
		{ID: "u3", TeamName: teamName},
	}

	users.
		On("GetByID", mock.Anything, authorID).
		Return(author, nil)

	users.
		On("ListActiveByTeam", mock.Anything, teamName).
		Return(members, nil)

	chooser.result = []string{"u2", "u3"}

	expected := domain.PullRequest{
		ID:                prID,
		Name:              prName,
		AuthorID:          authorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u2", "u3"},
	}

	prs.
		On("Create", mock.Anything, mock.MatchedBy(func(pr domain.PullRequest) bool {
			if pr.ID != expected.ID ||
				pr.Name != expected.Name ||
				pr.AuthorID != expected.AuthorID ||
				pr.Status != expected.Status {
				return false
			}
			return reflect.DeepEqual(pr.AssignedReviewers, expected.AssignedReviewers)
		})).
		Return(nil)

	got, err := uc.Create(ctx, prID, prName, authorID)
	if err != nil {
		t.Fatalf("Create() got err: %v, want nil", err)
	}

	if !chooser.called {
		t.Fatalf("Create() chooser should be called")
	}
	if chooser.limit != PrReviewersCount {
		t.Fatalf("Create() chooser.limit = %d, want %d", chooser.limit, PrReviewersCount)
	}
	for _, id := range chooser.candidates {
		if id == authorID {
			t.Fatalf("Create() candidates contain author id %q: %v", authorID, chooser.candidates)
		}
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("Create() got = %#v, want %#v", got, expected)
	}

	users.AssertExpectations(t)
	prs.AssertExpectations(t)
}

func TestUsecase_Create_GetAuthor_AppError(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	const (
		prID     = "pr-1"
		prName   = "Add feature"
		authorID = "user-404"
	)

	uc, users, prs, chooser := newPRUsecase()

	appErr := domain.NotFound("author not found")

	users.
		On("GetByID", mock.Anything, authorID).
		Return(domain.User{}, appErr)

	got, err := uc.Create(ctx, prID, prName, authorID)

	if !errors.Is(err, appErr) {
		t.Fatalf("Create() got err: %v, want appErr %v", err, appErr)
	}
	assertZeroPR(t, got)

	users.AssertNotCalled(t, "ListActiveByTeam", mock.Anything, mock.Anything)
	prs.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	if chooser.called {
		t.Fatalf("Create() chooser should not be called when author not found")
	}

	users.AssertExpectations(t)
	prs.AssertExpectations(t)
}

func TestUsecase_Create_PRCreate_AppError(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	const (
		prID     = "pr-1"
		prName   = "Add feature"
		authorID = "user-1"
		teamName = "backend"
	)

	uc, users, prs, chooser := newPRUsecase()

	author := domain.User{
		ID:       authorID,
		TeamName: teamName,
	}
	members := []domain.User{
		{ID: authorID, TeamName: teamName},
		{ID: "u2", TeamName: teamName},
	}

	appErr := domain.PRExists("already exists")

	users.
		On("GetByID", mock.Anything, authorID).
		Return(author, nil)

	users.
		On("ListActiveByTeam", mock.Anything, teamName).
		Return(members, nil)

	chooser.result = []string{"u2"}

	prs.
		On("Create", mock.Anything, mock.Anything).
		Return(appErr)

	got, err := uc.Create(ctx, prID, prName, authorID)

	if !errors.Is(err, appErr) {
		t.Fatalf("Create() got err: %v, want appErr %v", err, appErr)
	}
	assertZeroPR(t, got)

	users.AssertExpectations(t)
	prs.AssertExpectations(t)
}

func TestUsecase_Merge_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	const prID = "pr-1"

	uc, users, prs, _ := newPRUsecase()

	existing := domain.PullRequest{
		ID:       prID,
		Name:     "Feature",
		AuthorID: "user-1",
		Status:   domain.PRStatusOpen,
	}

	prs.
		On("GetByID", mock.Anything, prID).
		Return(existing, nil)

	prs.
		On("Update", mock.Anything, mock.MatchedBy(func(pr domain.PullRequest) bool {
			return pr.ID == existing.ID && pr.Status == domain.PRStatusMerged
		})).
		Return(nil)

	got, err := uc.Merge(ctx, prID)
	if err != nil {
		t.Fatalf("Merge() got err: %v, want nil", err)
	}

	if got.ID != existing.ID {
		t.Fatalf("Merge() got id %q, want %q", got.ID, existing.ID)
	}
	if got.Status != domain.PRStatusMerged {
		t.Fatalf("Merge() got status %q, want %q", got.Status, domain.PRStatusMerged)
	}

	prs.AssertExpectations(t)
	users.AssertExpectations(t)
}

func TestUsecase_Merge_AlreadyMerged(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	const prID = "pr-1"

	uc, users, prs, _ := newPRUsecase()

	existing := domain.PullRequest{
		ID:       prID,
		Name:     "Feature",
		AuthorID: "user-1",
		Status:   domain.PRStatusMerged,
	}

	prs.
		On("GetByID", mock.Anything, prID).
		Return(existing, nil)

	got, err := uc.Merge(ctx, prID)
	if err != nil {
		t.Fatalf("Merge() got err: %v, want nil", err)
	}

	prs.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)

	if !reflect.DeepEqual(got, existing) {
		t.Fatalf("Merge() got = %#v, want %#v", got, existing)
	}

	prs.AssertExpectations(t)
	users.AssertExpectations(t)
}

func TestUsecase_Merge_GetByID_AppError(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	const prID = "pr-404"

	uc, users, prs, _ := newPRUsecase()

	appErr := domain.NotFound("not found")

	prs.
		On("GetByID", mock.Anything, prID).
		Return(domain.PullRequest{}, appErr)

	got, err := uc.Merge(ctx, prID)

	if !errors.Is(err, appErr) {
		t.Fatalf("Merge() got err: %v, want appErr %v", err, appErr)
	}
	assertZeroPR(t, got)

	prs.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	prs.AssertExpectations(t)
	users.AssertExpectations(t)
}

func TestUsecase_Merge_Update_AppError(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	const prID = "pr-1"

	uc, users, prs, _ := newPRUsecase()

	existing := domain.PullRequest{
		ID:       prID,
		Name:     "Feature",
		AuthorID: "user-1",
		Status:   domain.PRStatusOpen,
	}

	appErr := domain.PRMerged("cannot update")

	prs.
		On("GetByID", mock.Anything, prID).
		Return(existing, nil)

	prs.
		On("Update", mock.Anything, mock.Anything).
		Return(appErr)

	got, err := uc.Merge(ctx, prID)

	if !errors.Is(err, appErr) {
		t.Fatalf("Merge() got err: %v, want appErr %v", err, appErr)
	}
	assertZeroPR(t, got)

	prs.AssertExpectations(t)
	users.AssertExpectations(t)
}
