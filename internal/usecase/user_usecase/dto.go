package usecase

import (
	"github.com/smthjapanese/avito-merch/internal/entity"
	"time"
)

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

type InventoryItemDTO struct {
	ItemID      int64     `json:"item_id"`
	ItemName    string    `json:"name"`
	Quantity    int64     `json:"quantity"`
	PurchasedAt time.Time `json:"purchased_at"`
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
