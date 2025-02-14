package usecase_user

import (
	"context"
	"errors"
	"fmt"
	"github.com/smthjapanese/avito-merch/internal/entity"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserUseCase struct {
	repo UserRepository
}

func NewUserUseCase(repo UserRepository) *UserUseCase {
	return &UserUseCase{
		repo: repo,
	}
}

// Register создает нового пользователя с указанным именем и паролем.
func (uc *UserUseCase) Register(ctx context.Context, username, password string) error {
	existingUser, err := uc.repo.GetUserByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return entity.ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
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
		return entity.ErrNegativeAmount
	}

	tx, err := uc.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	fromUser, err := tx.GetUserByID(ctx, fromUserID)
	if err != nil {
		return fmt.Errorf("failed to get sender: %w", err)
	}

	if fromUser.Coins < amount {
		return entity.ErrInsufficientFunds
	}

	toUser, err := tx.GetUserByUsername(ctx, toUsername)
	if err != nil {
		return fmt.Errorf("failed to get recipient: %w", err)
	}

	tr := &entity.Transaction{
		FromUserID: fromUserID,
		ToUserID:   toUser.ID,
		Amount:     amount,
		Type:       entity.TransactionTypeTransfer,
		CreatedAt:  time.Now(),
	}

	if err := tx.UpdateUserCoins(ctx, fromUserID, -amount); err != nil {
		return fmt.Errorf("failed to update sender balance: %w", err)
	}
	if err := tx.UpdateUserCoins(ctx, toUser.ID, amount); err != nil {
		return fmt.Errorf("failed to update recipient balance: %w", err)
	}

	if err := tx.CreateTransaction(ctx, tr); err != nil {
		return fmt.Errorf("failed to create transaction record: %w", err)
	}

	return tx.Commit()
}

func (uc *UserUseCase) BuyMerch(ctx context.Context, userID int64, merchName string) error {
	merch, err := uc.repo.GetMerchByName(ctx, merchName)
	if err != nil {
		return fmt.Errorf("failed to get merch item: %w", err)
	}

	user, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.Coins < merch.Price {
		return entity.ErrInsufficientFunds
	}

	tr := &entity.Transaction{
		FromUserID: userID,
		Amount:     merch.Price,
		Type:       entity.TransactionTypePurchase,
		ItemID:     &merch.ID,
		CreatedAt:  time.Now(),
	}

	if err := uc.repo.UpdateUserCoins(ctx, userID, -merch.Price); err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	if err := uc.repo.AddToInventory(ctx, userID, merch.ID); err != nil {
		return fmt.Errorf("failed to add item to inventory: %w", err)
	}

	return uc.repo.CreateTransaction(ctx, tr)
}

// GetInfo возвращает полную информацию о пользователе: баланс, инвентарь и историю транзакций.rm -
func (uc *UserUseCase) GetInfo(ctx context.Context, userID int64) (*entity.InfoResponse, error) {
	user, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return nil, entity.ErrUserNotFound
		}
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
		return nil, fmt.Errorf("failed to get user inventory: %w", err)
	}

	for _, item := range inventory {
		response.Inventory = append(response.Inventory, entity.InventoryItem{
			Type:     "merch",
			Quantity: item.Quantity,
		})
	}

	transactions, err := uc.repo.GetUserTransactions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user transactions: %w", err)
	}
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
