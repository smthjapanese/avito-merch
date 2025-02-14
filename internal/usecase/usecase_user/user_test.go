package usecase_user

import (
	"context"
	"github.com/smthjapanese/avito-merch/internal/entity"
	"github.com/smthjapanese/avito-merch/internal/usecase/usecase_user/mocks"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

// Тесты
func TestUserUseCase_SendCoins(t *testing.T) {
	repo := mocks.NewMockUserRepository()
	uc := NewUserUseCase(repo)
	ctx := context.Background()

	// Создаем тестовых пользователей
	sender := &entity.User{
		Username: "sender",
		Coins:    1000,
	}
	receiver := &entity.User{
		Username: "receiver",
		Coins:    0,
	}

	repo.CreateUser(ctx, sender)
	repo.CreateUser(ctx, receiver)

	tests := []struct {
		name      string
		fromID    int64
		toUser    string
		amount    int64
		wantError bool
	}{
		{
			name:      "Valid transfer",
			fromID:    1,
			toUser:    "receiver",
			amount:    500,
			wantError: false,
		},
		{
			name:      "Insufficient funds",
			fromID:    1,
			toUser:    "receiver",
			amount:    2000,
			wantError: true,
		},
		{
			name:      "Invalid amount",
			fromID:    1,
			toUser:    "receiver",
			amount:    -100,
			wantError: true,
		},
		{
			name:      "User not found",
			fromID:    999,
			toUser:    "receiver",
			amount:    100,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.SendCoins(ctx, tt.fromID, tt.toUser, tt.amount)
			if (err != nil) != tt.wantError {
				t.Errorf("SendCoins() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestUserUseCase_BuyMerch(t *testing.T) {
	repo := mock.NewMockRepository()
	uc := NewUserUseCase(repo)
	ctx := context.Background()

	// Создаем тестового пользователя
	user := &entity.User{
		Username: "testuser",
		Coins:    1000,
	}
	repo.CreateUser(ctx, user)

	// Создаем тестовые товары
	repo.merchItems["tshirt"] = &entity.MerchItem{
		ID:    1,
		Name:  "tshirt",
		Price: 500,
	}
	repo.merchItems["hoodie"] = &entity.MerchItem{
		ID:    2,
		Name:  "hoodie",
		Price: 1500,
	}

	tests := []struct {
		name      string
		userID    int64
		merchName string
		wantError bool
	}{
		{
			name:      "Valid purchase",
			userID:    1,
			merchName: "tshirt",
			wantError: false,
		},
		{
			name:      "Insufficient funds",
			userID:    1,
			merchName: "hoodie",
			wantError: true,
		},
		{
			name:      "Merch not found",
			userID:    1,
			merchName: "nonexistent",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.BuyMerch(ctx, tt.userID, tt.merchName)
			if (err != nil) != tt.wantError {
				t.Errorf("BuyMerch() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestUserUseCase_GetInfo(t *testing.T) {
	repo := NewMockRepository()
	uc := NewUserUseCase(repo)
	ctx := context.Background()

	// Создаем двух тестовых пользователей
	user1 := &entity.User{
		Username: "testuser1",
		Coins:    1000,
	}
	user2 := &entity.User{
		Username: "testuser2",
		Coins:    500,
	}
	// Создаем пользователей в репозитории
	if err := repo.CreateUser(ctx, user1); err != nil {
		t.Fatalf("Failed to create test user1: %v", err)
	}
	if err := repo.CreateUser(ctx, user2); err != nil {
		t.Fatalf("Failed to create test user2: %v", err)
	}

	// Добавляем тестовый мерч
	merch := &entity.MerchItem{
		ID:    1,
		Name:  "tshirt",
		Price: 100,
	}
	repo.merchItems["tshirt"] = merch

	// Добавляем тестовые транзакции
	tr1 := &entity.Transaction{
		FromUserID: user1.ID, // ID = 1
		ToUserID:   user2.ID, // ID = 2
		Amount:     100,
		Type:       entity.TransactionTypeTransfer,
		CreatedAt:  time.Now(),
	}
	if err := repo.CreateTransaction(ctx, tr1); err != nil {
		t.Fatalf("Failed to create test transaction: %v", err)
	}

	// Добавляем тестовый инвентарь
	inv := &entity.UserInventory{
		UserID:   user1.ID,
		ItemID:   merch.ID,
		Quantity: 1,
	}
	repo.inventory = append(repo.inventory, inv)

	// Тестируем получение информации
	t.Run("Get user info", func(t *testing.T) {
		info, err := uc.GetInfo(ctx, user1.ID)
		if err != nil {
			t.Errorf("GetInfo() error = %v", err)
			return
		}

		// Проверяем баланс
		if info.Coins != 1000 {
			t.Errorf("Expected coins = 1000, got %v", info.Coins)
		}

		// Проверяем инвентарь
		if len(info.Inventory) != 1 {
			t.Errorf("Expected 1 inventory item, got %v", len(info.Inventory))
		}

		// Проверяем историю транзакций
		if len(info.CoinHistory.Sent) != 1 {
			t.Errorf("Expected 1 sent transaction, got %v", len(info.CoinHistory.Sent))
		}
	})

	t.Run("Get info for non-existent user", func(t *testing.T) {
		_, err := uc.GetInfo(ctx, 999)
		if err == nil {
			t.Error("Expected error for non-existent user, got nil")
		}
	})
}
