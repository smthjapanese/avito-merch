package usecase

import (
	"context"
	"fmt"
	"github.com/smthjapanese/avito-merch/internal/entity"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserUseCase struct {
	userRepo  UserRepository
	txRepo    TransactionRepository
	invRepo   InventoryRepository
	merchRepo MerchRepository
}

func NewUserUseCase(
	userRepo UserRepository,
	txRepo TransactionRepository,
	invRepo InventoryRepository,
	merchRepo MerchRepository,
) UserUseCase {
	return UserUseCase{
		userRepo:  userRepo,
		txRepo:    txRepo,
		invRepo:   invRepo,
		merchRepo: merchRepo,
	}
}

func (uc *UserUseCase) Register(ctx context.Context, username, password string) (string, error) {
	existingUser, err := uc.userRepo.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return "", entity.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &entity.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Coins:        entity.InitialBalance,
		CreatedAt:    time.Now(),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return "", err
	}

	token := generateDummyToken(user.ID, username)
	return token, nil
}

func (uc *UserUseCase) GetProfile(ctx context.Context, userID int64) (UserProfileDTO, error) {
	// Get user info
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return UserProfileDTO{}, err
	}
	if user == nil {
		return UserProfileDTO{}, entity.ErrUserNotFound
	}

	inventory, err := uc.invRepo.GetByUserID(ctx, userID)
	if err != nil {
		return UserProfileDTO{}, err
	}

	// Get transaction history
	transactions, err := uc.txRepo.GetByUserID(ctx, userID)
	if err != nil {
		return UserProfileDTO{}, err
	}

	// Convert inventory to DTO
	inventoryDTO := make([]InventoryItemDTO, 0, len(inventory))
	for _, item := range inventory {
		merchItem, err := uc.merchRepo.GetByID(ctx, item.ItemID)
		if err != nil {
			continue // Skip items that can't be found
		}
		inventoryDTO = append(inventoryDTO, InventoryItemDTO{
			ItemID:      item.ItemID,
			ItemName:    merchItem.Name,
			Quantity:    item.Quantity,
			PurchasedAt: item.PurchasedAt,
		})
	}

	history := uc.processTransactionHistory(ctx, transactions, userID)

	return UserProfileDTO{
		User: UserDTO{
			ID:        user.ID,
			Username:  user.Username,
			Coins:     user.Coins,
			CreatedAt: user.CreatedAt,
		},
		Inventory: inventoryDTO,
		History:   history,
	}, nil
}

// Helper functions

func generateDummyToken(userID int64, username string) string {
	return fmt.Sprintf("dummy_token_%d_%s", userID, username)
}

func (uc *UserUseCase) processTransactionHistory(ctx context.Context, transactions []entity.Transaction, userID int64) TransactionHistory {
	received := make([]TransactionInfo, 0)
	sent := make([]TransactionInfo, 0)

	for _, tx := range transactions {
		var info TransactionInfo
		var otherUserName string

		if tx.ToUserID == userID {
			// Received transaction
			fromUser, err := uc.userRepo.GetByID(ctx, tx.FromUserID)
			if err != nil {
				continue
			}
			otherUserName = fromUser.Username
			info = TransactionInfo{
				ID:        tx.ID,
				User:      otherUserName,
				Amount:    tx.Amount,
				Type:      tx.Type,
				CreatedAt: tx.CreatedAt,
			}
			received = append(received, info)
		} else {
			// Sent transaction
			toUser, err := uc.userRepo.GetByID(ctx, tx.ToUserID)
			if err != nil {
				continue
			}
			otherUserName = toUser.Username
			info = TransactionInfo{
				ID:        tx.ID,
				User:      otherUserName,
				Amount:    tx.Amount,
				Type:      tx.Type,
				CreatedAt: tx.CreatedAt,
			}
			sent = append(sent, info)
		}

		// Add item name for purchase transactions
		if tx.Type == entity.TransactionTypePurchase && tx.ItemID != nil {
			merchItem, err := uc.merchRepo.GetByID(ctx, *tx.ItemID)
			if err == nil {
				itemName := merchItem.Name
				info.ItemName = &itemName
			}
		}
	}
	return TransactionHistory{
		Received: received,
		Sent:     sent,
	}
}
