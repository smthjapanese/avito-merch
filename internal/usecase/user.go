package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/evrone/go-clean-template/internal/entity"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserUseCase struct {
	repo Repository
}

func NewUserUseCase(repo Repository) *UserUseCase {
	return &UserUseCase{
		repo: repo,
	}
}

// Register создает нового пользователя с указанным именем и паролем.
func (uc *UserUseCase) Register(ctx context.Context, username, password string) error {
	existingUser, err := uc.repo.GetUserByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &entity.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Coins:        entity.InitialBalance,
		CreatedAt:    time.Now(),
	}

	return uc.repo.CreateUser(ctx, user)
}

// SendCoins переводит указанное количество монет от одного пользователя к другому.
func (uc *UserUseCase) SendCoins(ctx context.Context, fromUserID int64, toUsername string, amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	fromUser, err := uc.repo.GetUserByID(ctx, fromUserID)
	if err != nil {
		return err
	}

	if fromUser.Coins < amount {
		return errors.New("insufficient funds")
	}

	toUser, err := uc.repo.GetUserByUsername(ctx, toUsername)
	if err != nil {
		return err
	}

	tr := &entity.Transaction{
		FromUserID: fromUserID,
		ToUserID:   toUser.ID,
		Amount:     amount,
		Type:       entity.TransactionTypeTransfer,
		CreatedAt:  time.Now(),
	}

	// Атомарное обновление балансов обоих пользователей
	if err := uc.repo.UpdateUserCoins(ctx, fromUserID, -amount); err != nil {
		return err
	}
	if err := uc.repo.UpdateUserCoins(ctx, toUser.ID, amount); err != nil {
		return err
	}

	return uc.repo.CreateTransaction(ctx, tr)
}

// BuyMerch обрабатывает покупку мерча пользователем.
func (uc *UserUseCase) BuyMerch(ctx context.Context, userID int64, merchName string) error {
	merch, err := uc.repo.GetMerchByName(ctx, merchName)
	if err != nil {
		return err
	}

	user, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Coins < merch.Price {
		return errors.New("insufficient funds")
	}

	tr := &entity.Transaction{
		FromUserID: userID,
		Amount:     merch.Price,
		Type:       entity.TransactionTypePurchase,
		ItemID:     &merch.ID,
		CreatedAt:  time.Now(),
	}

	if err := uc.repo.UpdateUserCoins(ctx, userID, -merch.Price); err != nil {
		return err
	}

	if err := uc.repo.AddToInventory(ctx, userID, merch.ID); err != nil {
		return err
	}

	return uc.repo.CreateTransaction(ctx, tr)
}

// GetInfo возвращает полную информацию о пользователе: баланс, инвентарь и историю транзакций.rm -
func (uc *UserUseCase) GetInfo(ctx context.Context, userID int64) (*entity.InfoResponse, error) {
	user, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	response := &entity.InfoResponse{
		Coins:     user.Coins,
		Inventory: make([]entity.InventoryItem, 0),
		CoinHistory: entity.CoinHistory{
			Received: make([]entity.CoinReceived, 0),
			Sent:     make([]entity.CoinSent, 0),
		},
	}

	inventory, err := uc.repo.GetUserInventory(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	for _, item := range inventory {
		response.Inventory = append(response.Inventory, entity.InventoryItem{
			Type:     "merch",
			Quantity: item.Quantity,
		})
	}

	transactions, err := uc.repo.GetUserTransactions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Группировка транзакций по типу: входящие и исходящие
	for _, tr := range transactions {
		if tr.ToUserID == userID && tr.Type == entity.TransactionTypeTransfer {
			fromUser, err := uc.repo.GetUserByID(ctx, tr.FromUserID)
			if err != nil {
				continue
			}
			response.CoinHistory.Received = append(response.CoinHistory.Received, entity.CoinReceived{
				FromUser: fromUser.Username,
				Amount:   tr.Amount,
			})
		} else if tr.FromUserID == userID && tr.Type == entity.TransactionTypeTransfer {
			toUser, err := uc.repo.GetUserByID(ctx, tr.ToUserID)
			if err != nil {
				continue
			}
			response.CoinHistory.Sent = append(response.CoinHistory.Sent, entity.CoinSent{
				ToUser: toUser.Username,
				Amount: tr.Amount,
			})
		}
	}

	return response, nil
}
