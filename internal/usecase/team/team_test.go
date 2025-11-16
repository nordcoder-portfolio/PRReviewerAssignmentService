package team

import (
	"avito_test/internal/domain"
	"avito_test/internal/repo/postgres"
	"avito_test/mocks"
	"context"
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

type txMock struct {
	err    error
	called bool
}

func (m *txMock) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	m.called = true
	if m.err != nil {
		return m.err
	}
	return fn(ctx)
}

func (m *txMock) Querier(ctx context.Context) postgres.Querier {
	return nil
}

func assertZeroTeam(t *testing.T, got domain.Team) {
	t.Helper()
	require.Equal(t, domain.Team{}, got, "expected empty team here")
}

func TestCreateTeam(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	appErrCreate := domain.NotFound("team already exists")
	appErrUpsert := domain.NotFound("failed to upsert")
	repoErr := errors.New("db error")

	input := domain.Team{
		Name: "backend",
		Members: []domain.User{
			{ID: "u1"},
			{ID: "u2"},
		},
	}

	type want struct {
		err              error
		emptyTeam        bool
		expectUpsertMany bool
	}

	tests := []struct {
		name      string
		teamErr   error
		upsertErr error
		want      want
	}{
		{
			name:      "ok",
			teamErr:   nil,
			upsertErr: nil,
			want: want{
				err:              nil,
				emptyTeam:        false,
				expectUpsertMany: true,
			},
		},
		{
			name:      "team create app error",
			teamErr:   appErrCreate,
			upsertErr: nil, // won't be called
			want: want{
				err:              appErrCreate,
				emptyTeam:        true,
				expectUpsertMany: false,
			},
		},
		{
			name:      "team create repo error",
			teamErr:   repoErr,
			upsertErr: nil, // won't be called
			want: want{
				err:              repoErr,
				emptyTeam:        true,
				expectUpsertMany: false,
			},
		},
		{
			name:      "upsert many app error",
			teamErr:   nil,
			upsertErr: appErrUpsert,
			want: want{
				err:              appErrUpsert,
				emptyTeam:        true,
				expectUpsertMany: true,
			},
		},
		{
			name:      "upsert many repo error",
			teamErr:   nil,
			upsertErr: repoErr,
			want: want{
				err:              repoErr,
				emptyTeam:        true,
				expectUpsertMany: true,
			},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			teamsRepo := &mocks.TeamRepository{}
			usersRepo := &mocks.UserRepository{}
			tx := &txMock{}

			teamsRepo.
				On("Create", mock.Anything, input.Name).
				Return(tc.teamErr)

			if tc.want.expectUpsertMany {
				// здесь нам важно только то, что TeamName проставился,
				// сами ID нам пофиг
				usersRepo.
					On(
						"UpsertMany",
						mock.Anything,
						input.Name,
						mock.MatchedBy(func(users []domain.User) bool {
							if len(users) != len(input.Members) {
								return false
							}
							for _, u := range users {
								if u.TeamName != input.Name {
									return false
								}
							}
							return true
						}),
					).
					Return(tc.upsertErr)
			}

			uc := &usecase{
				teams:  teamsRepo,
				users:  usersRepo,
				tx:     tx,
				logger: newTestLogger(),
			}

			got, err := uc.CreateTeam(ctx, input)

			if tc.want.err != nil {
				require.Error(t, err, "expected error here")
				require.ErrorIs(t, err, tc.want.err, "got wrong error")
				assertZeroTeam(t, got)
			} else {
				require.NoError(t, err, "did not expect error here")
				require.Equal(t, input.Name, got.Name, "team name mismatch")
				require.Len(t, got.Members, len(input.Members), "members count mismatch")
				for i, u := range got.Members {
					require.Equal(t, input.Name, u.TeamName, "wrong team for member %d", i)
				}
			}

			require.True(t, tx.called, "tx should be used here")

			if !tc.want.expectUpsertMany {
				usersRepo.AssertNotCalled(t, "UpsertMany", mock.Anything, mock.Anything, mock.Anything)
			}

			teamsRepo.AssertExpectations(t)
			usersRepo.AssertExpectations(t)
		})
	}
}

func TestGetTeam(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	const teamName = "backend"

	appErr := domain.NotFound("team not found")
	repoErr := errors.New("db error")

	expectedTeam := domain.Team{
		Name: teamName,
		Members: []domain.User{
			{ID: "u1", TeamName: teamName},
		},
	}

	type want struct {
		team domain.Team
		err  error
	}

	tests := []struct {
		name     string
		repoTeam domain.Team
		repoErr  error
		want     want
	}{
		{
			name:     "ok",
			repoTeam: expectedTeam,
			repoErr:  nil,
			want: want{
				team: expectedTeam,
				err:  nil,
			},
		},
		{
			name:     "app error",
			repoTeam: domain.Team{},
			repoErr:  appErr,
			want: want{
				team: domain.Team{},
				err:  appErr,
			},
		},
		{
			name:     "repo error",
			repoTeam: domain.Team{},
			repoErr:  repoErr,
			want: want{
				team: domain.Team{},
				err:  repoErr,
			},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			teamsRepo := &mocks.TeamRepository{}
			usersRepo := &mocks.UserRepository{}

			teamsRepo.
				On("GetByName", mock.Anything, teamName).
				Return(tc.repoTeam, tc.repoErr)

			uc := &usecase{
				teams:  teamsRepo,
				users:  usersRepo,
				logger: newTestLogger(),
			}

			got, err := uc.GetTeam(ctx, teamName)

			if tc.want.err != nil {
				require.Error(t, err, "expected error here")
				require.ErrorIs(t, err, tc.want.err, "got wrong error")
				assertZeroTeam(t, got)
			} else {
				require.NoError(t, err, "did not expect error here")
				require.Equal(t, tc.want.team, got, "team mismatch")
			}

			teamsRepo.AssertExpectations(t)
			usersRepo.AssertExpectations(t)
		})
	}
}
