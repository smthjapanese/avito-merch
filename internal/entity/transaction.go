package entity

import "time"

type TransactionType string

const (
	TransactionTypeTransfer TransactionType = "transfer"
	TransactionTypePurchase TransactionType = "purchase"
)

type Transaction struct {
	ID         int64           `json:"id" db:"id"`
	FromUserID int64           `json:"from_user_id" db:"from_user_id"`
	ToUserID   int64           `json:"to_user_id" db:"to_user_id"`
	Amount     int64           `json:"amount" db:"amount"`
	Type       TransactionType `json:"type" db:"type"`
	ItemID     *int64          `json:"item_id,omitempty" db:"item_id"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}
