package entity

import "time"

type UserInventory struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	ItemID      int64     `json:"item_id" db:"item_id"`
	Quantity    int64     `json:"quantity" db:"quantity"`
	PurchasedAt time.Time `json:"purchased_at" db:"purchased_at"`
}
