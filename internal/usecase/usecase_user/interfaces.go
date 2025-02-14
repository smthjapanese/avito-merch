package usecase_user

import (
	"context"
	"github.com/smthjapanese/avito-merch/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks/mocks.go -package=mocks

type UserRepository interface {
	BeginTx(ctx context.Context) (entity.Transaction, error)
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	GetUserByID(ctx context.Context, id int64) (*entity.User, error)
	UpdateUserCoins(ctx context.Context, userID int64, amount int64) error
	GetMerchByName(ctx context.Context, name string) (*entity.MerchItem, error)
	AddToInventory(ctx context.Context, userID, itemID int64) error
	CreateTransaction(ctx context.Context, tr *entity.Transaction) error
	GetUserTransactions(ctx context.Context, userID int64) ([]*entity.Transaction, error)
	GetUserInventory(ctx context.Context, userID int64) ([]*entity.UserInventory, error)
}