package usecase

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/smthjapanese/avito-merch/internal/entity"
	"github.com/smthjapanese/avito-merch/internal/usecase/merch_usecase/mocks"
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

func TestMerchUseCase_ListAvailable(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	merchRepo := mocks.NewMockMerchRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	invRepo := mocks.NewMockInventoryRepository(ctrl)
	txRepo := mocks.NewMockTransactionRepository(ctrl)
	dbTransactor := mocks.NewMockDBTransactor(ctrl)

	uc := NewMerchUseCase(merchRepo, userRepo, invRepo, txRepo, dbTransactor)

	testTime := time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC)
	testItems := []entity.MerchItem{
		{
			ID:        1,
			Name:      "Test Item 1",
			Price:     100,
			CreatedAt: testTime,
		},
		{
			ID:        2,
			Name:      "Test Item 2",
			Price:     200,
			CreatedAt: testTime,
		},
	}

	tests := []test{
		{
			name: "success",
			mock: func() {
				merchRepo.EXPECT().
					List(gomock.Any()).
					Return(testItems, nil)
			},
			res: []MerchItemDTO{
				{
					ID:    1,
					Name:  "Test Item 1",
					Price: 100,
				},
				{
					ID:    2,
					Name:  "Test Item 2",
					Price: 200,
				},
			},
			err: nil,
		},
		{
			name: "repository error",
			mock: func() {
				merchRepo.EXPECT().
					List(gomock.Any()).
					Return(nil, entity.ErrTransactionFailed)
			},
			res: nil,
			err: entity.ErrTransactionFailed,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()

			items, err := uc.ListAvailable(context.Background())

			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.res, items)
			}
		})
	}
}

func TestMerchUseCase_BuyItem(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	merchRepo := mocks.NewMockMerchRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	invRepo := mocks.NewMockInventoryRepository(ctrl)
	txRepo := mocks.NewMockTransactionRepository(ctrl)
	dbTransactor := mocks.NewMockDBTransactor(ctrl)

	uc := NewMerchUseCase(merchRepo, userRepo, invRepo, txRepo, dbTransactor)

	testTime := time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC)
	userID := int64(1)
	itemName := "Test Item"

	testItem := entity.MerchItem{
		ID:        1,
		Name:      itemName,
		Price:     100,
		CreatedAt: testTime,
	}

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
			Quantity:    1,
			PurchasedAt: testTime,
		},
	}

	tests := []test{
		{
			name: "success_first_purchase",
			mock: func() {
				var updatedUser *entity.User
				var createdInventory entity.UserInventory

				dbTransactor.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})

				merchRepo.EXPECT().
					GetByName(gomock.Any(), itemName).
					Return(testItem, nil)

				userRepo.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(testUser, nil)

				userRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *entity.User) error {
						updatedUser = user
						return nil
					})

				invRepo.EXPECT().
					GetByUserID(gomock.Any(), userID).
					Return(nil, nil)

				invRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inventory entity.UserInventory) error {
						createdInventory = inventory
						return nil
					})

				txRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)

				// Проверяем обновленный баланс после выполнения операции
				t.Cleanup(func() {
					require.NotNil(t, updatedUser)
					require.Equal(t, testUser.Coins-testItem.Price, updatedUser.Coins)

					require.Equal(t, userID, createdInventory.UserID)
					require.Equal(t, testItem.ID, createdInventory.ItemID)
					require.Equal(t, int64(1), createdInventory.Quantity)
				})
			},
			err: nil,
		},
		{
			name: "success_repeat_purchase",
			mock: func() {
				var updatedUser *entity.User
				var updatedInventory entity.UserInventory

				dbTransactor.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})

				merchRepo.EXPECT().
					GetByName(gomock.Any(), itemName).
					Return(testItem, nil)

				userRepo.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(testUser, nil)

				userRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *entity.User) error {
						updatedUser = user
						return nil
					})

				invRepo.EXPECT().
					GetByUserID(gomock.Any(), userID).
					Return(testInventory, nil)

				invRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, inventory entity.UserInventory) error {
						updatedInventory = inventory
						return nil
					})

				txRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)

				t.Cleanup(func() {
					require.NotNil(t, updatedUser)
					require.Equal(t, testUser.Coins-testItem.Price, updatedUser.Coins)
					require.Equal(t, int64(2), updatedInventory.Quantity)
				})
			},
			err: nil,
		},
		{
			name: "item_not_found",
			mock: func() {
				dbTransactor.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})

				merchRepo.EXPECT().
					GetByName(gomock.Any(), itemName).
					Return(entity.MerchItem{}, entity.ErrMerchNotFound)
			},
			err: entity.ErrMerchNotFound,
		},
		{
			name: "insufficient_funds",
			mock: func() {
				dbTransactor.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})

				merchRepo.EXPECT().
					GetByName(gomock.Any(), itemName).
					Return(testItem, nil)

				userRepo.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(&entity.User{
						ID:        userID,
						Username:  "testuser",
						Coins:     50, // Less than item price
						CreatedAt: testTime,
					}, nil)
			},
			err: entity.ErrInsufficientFunds,
		},
		{
			name: "transaction_failed_on_user_update",
			mock: func() {
				dbTransactor.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})

				merchRepo.EXPECT().
					GetByName(gomock.Any(), itemName).
					Return(testItem, nil)

				userRepo.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(testUser, nil)

				userRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(entity.ErrTransactionFailed)
			},
			err: entity.ErrTransactionFailed,
		},
		{
			name: "transaction_failed_on_inventory_update",
			mock: func() {
				dbTransactor.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})

				merchRepo.EXPECT().
					GetByName(gomock.Any(), itemName).
					Return(testItem, nil)

				userRepo.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(testUser, nil)

				userRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil)

				invRepo.EXPECT().
					GetByUserID(gomock.Any(), userID).
					Return(testInventory, nil)

				invRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(entity.ErrTransactionFailed)
			},
			err: entity.ErrTransactionFailed,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			err := uc.BuyItem(context.Background(), userID, itemName)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
