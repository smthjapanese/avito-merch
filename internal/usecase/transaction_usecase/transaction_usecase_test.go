package transaction_usecase

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/smthjapanese/avito-merch/internal/entity"
	"github.com/smthjapanese/avito-merch/internal/usecase/transaction_usecase/mocks"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type test struct {
	name   string
	mock   func()
	fromID int64
	toID   int64
	amount int64
	res    interface{}
	err    error
}

func TestCreateTransfer(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	txRepo := mocks.NewMockRepository(ctrl)
	dbTx := mocks.NewMockDBTransactor(ctrl)

	uc := NewTransactionUC(userRepo, txRepo, dbTx)

	fromUser := &entity.User{
		ID:       1,
		Username: "sender",
		Coins:    1000,
	}

	toUser := &entity.User{
		ID:       2,
		Username: "receiver",
		Coins:    500,
	}

	tests := []test{
		{
			name:   "success",
			fromID: 1,
			toID:   2,
			amount: 500,
			mock: func() {
				dbTx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})

				userRepo.EXPECT().
					GetByID(gomock.Any(), int64(1)).
					Return(fromUser, nil)

				userRepo.EXPECT().
					GetByID(gomock.Any(), int64(2)).
					Return(toUser, nil)

				userRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *entity.User) error {
						require.Equal(t, int64(500), user.Coins) // 1000 - 500
						return nil
					})

				userRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *entity.User) error {
						require.Equal(t, int64(1000), user.Coins) // 500 + 500
						return nil
					})

				txRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, tx entity.Transaction) error {
						require.Equal(t, int64(1), tx.FromUserID)
						require.Equal(t, int64(2), tx.ToUserID)
						require.Equal(t, int64(500), tx.Amount)
						require.Equal(t, entity.TransactionTypeTransfer, tx.Type)
						return nil
					})
			},
			res: nil,
			err: nil,
		},
		{
			name:   "insufficient funds",
			fromID: 1,
			toID:   2,
			amount: 2000, // больше чем есть на балансе
			mock: func() {
				dbTx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})

				userRepo.EXPECT().
					GetByID(gomock.Any(), int64(1)).
					Return(fromUser, nil)

				userRepo.EXPECT().
					GetByID(gomock.Any(), int64(2)).
					Return(toUser, nil)
			},
			res: nil,
			err: entity.ErrInsufficientFunds,
		},
		{
			name:   "sender not found",
			fromID: 999,
			toID:   2,
			amount: 500,
			mock: func() {
				dbTx.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})

				userRepo.EXPECT().
					GetByID(gomock.Any(), int64(999)).
					Return(nil, entity.ErrUserNotFound)
			},
			res: nil,
			err: entity.ErrTransactionFailed,
		},
		{
			name:   "negative amount",
			fromID: 1,
			toID:   2,
			amount: -100,
			mock:   func() {}, // Для отрицательной суммы даже не дойдет до транзакции
			res:    nil,
			err:    entity.ErrNegativeAmount,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.mock()

			err := uc.CreateTransfer(context.Background(), tc.fromID, tc.toID, tc.amount)

			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetUserHistory(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	txRepo := mocks.NewMockRepository(ctrl)
	dbTx := mocks.NewMockDBTransactor(ctrl)

	uc := NewTransactionUC(userRepo, txRepo, dbTx)
	testTime := time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC)
	userID := int64(1)

	testUser := &entity.User{
		ID:       userID,
		Username: "testuser",
		Coins:    1000,
	}

	senderUser := &entity.User{
		ID:       2,
		Username: "sender",
	}

	receiverUser := &entity.User{
		ID:       3,
		Username: "receiver",
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

	tests := []test{
		{
			name: "success",
			mock: func() {
				userRepo.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(testUser, nil)

				txRepo.EXPECT().
					GetByUserID(gomock.Any(), userID).
					Return(testTransactions, nil)

				userRepo.EXPECT().
					GetByID(gomock.Any(), int64(2)). // sender
					Return(senderUser, nil)

				userRepo.EXPECT().
					GetByID(gomock.Any(), int64(3)). // receiver
					Return(receiverUser, nil)
			},
			res: &TransactionHistory{
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
			err: nil,
		},
		{
			name: "user not found",
			mock: func() {
				userRepo.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(nil, entity.ErrUserNotFound)
			},
			res: nil,
			err: entity.ErrUserNotFound,
		},
		{
			name: "transaction fetch error",
			mock: func() {
				userRepo.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(testUser, nil)

				txRepo.EXPECT().
					GetByUserID(gomock.Any(), userID).
					Return(nil, entity.ErrTransactionFailed)
			},
			res: nil,
			err: entity.ErrTransactionFailed,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.mock()

			history, err := uc.GetUserHistory(context.Background(), userID)

			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.res, history)
			}
		})
	}
}
