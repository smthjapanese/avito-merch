package entity

import "time"

const (
	InitialBalance int64 = 1000
)

type User struct {
	ID           int64     `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Coins        int64     `json:"coins" db:"coins"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
