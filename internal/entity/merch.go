package entity

import "time"

type MerchItem struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Price     int64     `json:"price" db:"price"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
