package usecase

import (
	"context"
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
