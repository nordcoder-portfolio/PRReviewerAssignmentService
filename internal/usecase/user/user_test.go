package user

import (
	"avito_test/internal/domain"
	mockpostgres "avito_test/mocks"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestUsecase_SetIsActive(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	appErr := domain.NotFound("user not found")
	repoErr := errors.New("db error")

	type args struct {
		userID   string
		isActive bool
	}

	tests := []struct {
		name     string
		args     args
		repoUser domain.User
		repoErr  error
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "ok",
			args: args{
				userID:   "user-123",
				isActive: true,
			},
			repoUser: domain.User{
				ID:       "user-123",
				TeamName: "team-1",
			},
			repoErr: nil,
			wantUser: domain.User{
				ID:       "user-123",
				TeamName: "team-1",
			},
			wantErr: nil,
		},
		{
			name: "app error",
			args: args{
				userID:   "user-404",
				isActive: false,
			},
			repoUser: domain.User{},
			repoErr:  appErr,
			wantUser: domain.User{},
			wantErr:  appErr,
		},
		{
			name: "repo error",
			args: args{
				userID:   "user-123",
				isActive: true,
			},
			repoUser: domain.User{},
			repoErr:  repoErr,
			wantUser: domain.User{},
			wantErr:  repoErr,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userRepoMock := &mockpostgres.UserRepository{}
			userRepoMock.
				On("SetIsActive", mock.Anything, tc.args.userID, tc.args.isActive).
				Return(tc.repoUser, tc.repoErr)

			prRepoMock := &mockpostgres.PRRepository{}

			uc := New(userRepoMock, prRepoMock, nil, newTestLogger())

			gotUser, err := uc.SetIsActive(ctx, tc.args.userID, tc.args.isActive)

			if tc.wantErr != nil {
				require.Error(t, err, "expected error here")
				require.ErrorIs(t, err, tc.wantErr, "wrong error type")
				require.Equal(t, tc.wantUser, gotUser, "user should be empty here")
			} else {
				require.NoError(t, err, "did not expect error here")
				require.Equal(t, tc.wantUser, gotUser, "user mismatch")
			}

			userRepoMock.AssertExpectations(t)
			prRepoMock.AssertExpectations(t)
		})
	}
}

func TestUsecase_GetReviewPRs(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	appErr := domain.NotFound("prs not found")
	repoErr := errors.New("db error")

	type args struct {
		userID string
	}

	tests := []struct {
		name    string
		args    args
		repoPRs []domain.PullRequestShort
		repoErr error
		wantPRs []domain.PullRequestShort
		wantErr error
	}{
		{
			name: "ok",
			args: args{
				userID: "user-123",
			},
			repoPRs: []domain.PullRequestShort{
				{ID: "pr-1", Name: "PR 1", AuthorID: "a1", Status: domain.PRStatusOpen},
				{ID: "pr-2", Name: "PR 2", AuthorID: "a2", Status: domain.PRStatusOpen},
			},
			repoErr: nil,
			wantPRs: []domain.PullRequestShort{
				{ID: "pr-1", Name: "PR 1", AuthorID: "a1", Status: domain.PRStatusOpen},
				{ID: "pr-2", Name: "PR 2", AuthorID: "a2", Status: domain.PRStatusOpen},
			},
			wantErr: nil,
		},
		{
			name: "app error",
			args: args{
				userID: "user-123",
			},
			repoPRs: nil,
			repoErr: appErr,
			wantPRs: nil,
			wantErr: appErr,
		},
		{
			name: "repo error",
			args: args{
				userID: "user-123",
			},
			repoPRs: nil,
			repoErr: repoErr,
			wantPRs: nil,
			wantErr: repoErr,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userRepoMock := &mockpostgres.UserRepository{}
			prRepoMock := &mockpostgres.PRRepository{}
			prRepoMock.
				On("ListByReviewer", mock.Anything, tc.args.userID).
				Return(tc.repoPRs, tc.repoErr)

			uc := New(userRepoMock, prRepoMock, nil, newTestLogger())

			gotPRs, err := uc.GetReviewPRs(ctx, tc.args.userID)

			if tc.wantErr != nil {
				require.Error(t, err, "expected error here")
				require.ErrorIs(t, err, tc.wantErr, "wrong error type")
				require.Nil(t, gotPRs, "expected nil prs on error")
			} else {
				require.NoError(t, err, "did not expect error here")
				require.Equal(t, tc.wantPRs, gotPRs, "prs list mismatch")
			}

			userRepoMock.AssertExpectations(t)
			prRepoMock.AssertExpectations(t)
		})
	}
}
