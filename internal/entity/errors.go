package entity

import "errors"

var (
	ErrNegativeAmount    = errors.New("amount must be positive")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrMerchNotFound     = errors.New("merch not found")
	ErrTransactionFailed = errors.New("transaction failed")
)
