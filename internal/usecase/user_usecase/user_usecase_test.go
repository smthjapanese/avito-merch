package usecase

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/smthjapanese/avito-merch/internal/entity"
	"github.com/smthjapanese/avito-merch/internal/usecase/user_usecase/mocks"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type test struct {
	name string
	mock func()
	res  interface{}
	err  error
}

func TestRegister(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	txRepo := mocks.NewMockTransactionRepository(ctrl)
	invRepo := mocks.NewMockInventoryRepository(ctrl)
	merchRepo := mocks.NewMockMerchRepository(ctrl)

	uc := NewUserUseCase(userRepo, txRepo, invRepo, merchRepo)

	tests := []test{
		{
			name: "success",
			mock: func() {
				userRepo.EXPECT().
					GetByUsername(gomock.Any(), "testuser").
					Return(nil, nil)

				userRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *entity.User) error {
						require.Equal(t, "testuser", user.Username)
						require.Equal(t, entity.InitialBalance, user.Coins)
						require.NotEmpty(t, user.PasswordHash)
						return nil
					})
			},
			res: "dummy_token_0_testuser",
			err: nil,
		},
		{
			name: "user already exists",
			mock: func() {
				userRepo.EXPECT().
					GetByUsername(gomock.Any(), "testuser").
					Return(&entity.User{Username: "testuser"}, nil)
			},
			res: "",
			err: entity.ErrUserAlreadyExists,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.mock()

			token, err := uc.Register(context.Background(), "testuser", "password123")

			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.res, token)
			}
		})
	}
}

func TestGetProfile(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	txRepo := mocks.NewMockTransactionRepository(ctrl)
	invRepo := mocks.NewMockInventoryRepository(ctrl)
	merchRepo := mocks.NewMockMerchRepository(ctrl)

	uc := NewUserUseCase(userRepo, txRepo, invRepo, merchRepo)

	testTime := time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC)
	userID := int64(1)
	testUser := &entity.User{
		ID:        userID,
		Username:  "testuser",
		Coins:     1000,
		CreatedAt: testTime,
	}

	testInventory := []entity.UserInventory{
		{
			ID:          1,
			UserID:      userID,
			ItemID:      1,
			Quantity:    2,
			PurchasedAt: testTime,
		},
	}

	testMerchItem := entity.MerchItem{
		ID:        1,
		Name:      "Test Item",
		Price:     100,
		CreatedAt: testTime,
	}

	testTransactions := []entity.Transaction{
		{
			ID:         1,
			FromUserID: 2,
			ToUserID:   userID,
			Amount:     500,
			Type:       entity.TransactionTypeTransfer,
			CreatedAt:  testTime,
		},
		{
			ID:         2,
			FromUserID: userID,
			ToUserID:   3,
			Amount:     200,
			Type:       entity.TransactionTypeTransfer,
			CreatedAt:  testTime,
		},
	}

	senderUser := &entity.User{
		ID:       2,
		Username: "sender",
	}

	receiverUser := &entity.User{
		ID:       3,
		Username: "receiver",
	}

	tests := []test{
		{
			name: "success",
			mock: func() {
				userRepo.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(testUser, nil)

				invRepo.EXPECT().
					GetByUserID(gomock.Any(), userID).
					Return(testInventory, nil)

				merchRepo.EXPECT().
					GetByID(gomock.Any(), int64(1)).
					Return(testMerchItem, nil)

				txRepo.EXPECT().
					GetByUserID(gomock.Any(), userID).
					Return(testTransactions, nil)

				userRepo.EXPECT().
					GetByID(gomock.Any(), int64(2)).
					Return(senderUser, nil)
				userRepo.EXPECT().
					GetByID(gomock.Any(), int64(3)).
					Return(receiverUser, nil)
			},
			res: UserProfileDTO{
				User: UserDTO{
					ID:        userID,
					Username:  "testuser",
					Coins:     1000,
					CreatedAt: testTime,
				},
				Inventory: []InventoryItemDTO{
					{
						ItemID:      1,
						ItemName:    "Test Item",
						Quantity:    2,
						PurchasedAt: testTime,
					},
				},
				History: TransactionHistory{
					Received: []TransactionInfo{
						{
							ID:        1,
							User:      "sender",
							Amount:    500,
							Type:      entity.TransactionTypeTransfer,
							CreatedAt: testTime,
						},
					},
					Sent: []TransactionInfo{
						{
							ID:        2,
							User:      "receiver",
							Amount:    200,
							Type:      entity.TransactionTypeTransfer,
							CreatedAt: testTime,
						},
					},
				},
			},
			err: nil,
		},
		{
			name: "user not found",
			mock: func() {
				userRepo.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(nil, entity.ErrUserNotFound)
			},
			res: UserProfileDTO{},
			err: entity.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.mock()

			profile, err := uc.GetProfile(context.Background(), userID)

			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.res, profile)
			}
		})
	}
}
