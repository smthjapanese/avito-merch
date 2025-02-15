package usecase

import (
	"context"
	"github.com/smthjapanese/avito-merch/internal/entity"
	"time"
)

type TransactionUseCase interface {
	CreateTransfer(ctx context.Context, fromUserID, toUserID int64, amount int64) error
	GetUserHistory(ctx context.Context, userID int64) (*TransactionHistory, error)
}

type MerchUseCase interface {
	ListAvailable(ctx context.Context) ([]MerchItemDTO, error)
	BuyItem(ctx context.Context, userID int64, itemName string) error
}

type UserUseCase interface {
	Register(ctx context.Context, username, password string) (string, error)
	GetProfile(ctx context.Context, userID int64) (UserProfileDTO, error)
}

type TransactionHistory struct {
	Received []TransactionInfo `json:"received"`
	Sent     []TransactionInfo `json:"sent"`
}

type TransactionInfo struct {
	ID        int64                  `json:"id"`
	User      string                 `json:"user"`
	Amount    int64                  `json:"amount"`
	Type      entity.TransactionType `json:"type"`
	ItemName  *string                `json:"item_name,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

type MerchItemDTO struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Price int64  `json:"price"`
}

type UserDTO struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Coins     int64     `json:"coins"`
	CreatedAt time.Time `json:"created_at"`
}

type UserProfileDTO struct {
	User      UserDTO            `json:"user"`
	Inventory []InventoryItemDTO `json:"inventory"`
	History   TransactionHistory `json:"history"`
}

type UserRegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6"`
}

type UserUpdateDTO struct {
	Password *string `json:"password,omitempty" validate:"omitempty,min=6"`
}

type InventoryItemDTO struct {
	ItemID      int64     `json:"item_id"`
	ItemName    string    `json:"name"`
	Quantity    int64     `json:"quantity"`
	PurchasedAt time.Time `json:"purchased_at"`
}
