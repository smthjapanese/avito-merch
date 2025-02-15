package transaction_usecase

import (
	"context"
	"github.com/smthjapanese/avito-merch/internal/entity"
)

type User interface {
	Register(ctx context.Context, username, password string) error
	Login(ctx context.Context, username, password string) (string, error)
	SendCoins(ctx context.Context, fromUserID int64, toUser string, amount int64) error
	BuyMerch(ctx context.Context, userID int64, merchName string) error
	BeginTx(ctx context.Context) (Transaction, error)
}

type Repository interface {
	BeginTx(ctx context.Context) (Transaction, error)
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
type Transaction interface {
	GetUserByID(ctx context.Context, id int64) (*entity.User, error)
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	UpdateUserCoins(ctx context.Context, userID int64, amount int64) error
	CreateTransaction(ctx context.Context, tr *entity.Transaction) error
	Commit() error
	Rollback() error
}
