package user_usecaseold

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/smthjapanese/avito-merch/internal/entity"
	"github.com/smthjapanese/avito-merch/internal/usecase/user_usecaseold/mocks"
	"github.com/stretchr/testify/require"
)

type mockClock struct {
	now time.Time
}

func (m mockClock) Now() time.Time {
	return m.now
}

type test struct {
	name string
	mock func()
	res  interface{}
	err  error
}

func userUseCase(t *testing.T) (*UserUseCase, *mocks.MockUserRepository, mockClock) {
	t.Helper()

	mockCtl := gomock.NewController(t)

	repo := mocks.NewMockUserRepository(mockCtl)
	clock := mockClock{now: time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC)}

	uc := NewUserUseCase(repo, clock)

	return &uc, repo, clock
}

func TestCreate(t *testing.T) {
	t.Parallel()

	uc, repo, clock := userUseCase(t)

	tests := []test{
		{
			name: "success",
			mock: func() {
				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, user entity.User) error {
						require.Equal(t, "testuser", user.Username)
						require.Equal(t, entity.InitialBalance, user.Coins)
						require.Equal(t, clock.Now(), user.CreatedAt)
						require.NotEmpty(t, user.PasswordHash)
						return nil
					})
			},
			res: entity.User{
				Username:  "testuser",
				Coins:     entity.InitialBalance,
				CreatedAt: clock.Now(),
			},
			err: nil,
		},
		{
			name: "user already exists",
			mock: func() {
				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(entity.ErrUserAlreadyExists)
			},
			res: entity.User{},
			err: entity.ErrUserAlreadyExists,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.mock()

			res, err := uc.Create(context.Background(), "testuser", "password123")

			if tc.err == nil {
				expectedUser := tc.res.(entity.User)
				require.Equal(t, expectedUser.Username, res.Username)
				require.Equal(t, expectedUser.Coins, res.Coins)
				require.Equal(t, expectedUser.CreatedAt, res.CreatedAt)
				require.NotEmpty(t, res.PasswordHash)
			} else {
				require.ErrorIs(t, err, tc.err)
			}
		})
	}
}

func TestGetByID(t *testing.T) {
	t.Parallel()

	uc, repo, _ := userUseCase(t)

	tests := []test{
		{
			name: "success",
			mock: func() {
				repo.EXPECT().
					GetByID(gomock.Any(), int64(1)).
					Return(entity.User{ID: 1, Username: "testuser"}, nil)
			},
			res: entity.User{ID: 1, Username: "testuser"},
			err: nil,
		},
		{
			name: "user not found",
			mock: func() {
				repo.EXPECT().
					GetByID(gomock.Any(), int64(1)).
					Return(entity.User{}, entity.ErrUserNotFound)
			},
			res: entity.User{},
			err: entity.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.mock()

			res, err := uc.GetByID(context.Background(), 1)

			require.Equal(t, tc.res, res)
			require.ErrorIs(t, err, tc.err)
		})
	}
}

func TestGetByUsername(t *testing.T) {
	t.Parallel()

	uc, repo, _ := userUseCase(t)

	tests := []test{
		{
			name: "success",
			mock: func() {
				repo.EXPECT().
					GetByUsername(gomock.Any(), "testuser").
					Return(entity.User{Username: "testuser"}, nil)
			},
			res: entity.User{Username: "testuser"},
			err: nil,
		},
		{
			name: "user not found",
			mock: func() {
				repo.EXPECT().
					GetByUsername(gomock.Any(), "testuser").
					Return(entity.User{}, entity.ErrUserNotFound)
			},
			res: entity.User{},
			err: entity.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.mock()

			res, err := uc.GetByUsername(context.Background(), "testuser")

			require.Equal(t, tc.res, res)
			require.ErrorIs(t, err, tc.err)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	uc, repo, _ := userUseCase(t)

	tests := []test{
		{
			name: "success",
			mock: func() {
				repo.EXPECT().
					Update(gomock.Any(), entity.User{ID: 1, Username: "testuser", Coins: 2000}).
					Return(nil)
			},
			err: nil,
		},
		{
			name: "user not found",
			mock: func() {
				repo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(entity.ErrUserNotFound)
			},
			err: entity.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.mock()

			err := uc.UpdateUser(context.Background(), entity.User{ID: 1, Username: "testuser", Coins: 2000})

			require.ErrorIs(t, err, tc.err)
		})
	}
}
