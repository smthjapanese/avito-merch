// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"
	"github.com/smthjapanese/avito-merch/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=usecase_test

type (
	// Translation -.
	Translation interface {
		Translate(context.Context, entity.Translation) (entity.Translation, error)
		History(context.Context) ([]entity.Translation, error)
	}

	// TranslationRepo -.
	TranslationRepo interface {
		Store(context.Context, entity.Translation) error
		GetHistory(context.Context) ([]entity.Translation, error)
	}

	// TranslationWebAPI -.
	TranslationWebAPI interface {
		Translate(entity.Translation) (entity.Translation, error)
	}
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
type DBTransactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type TransactionRepository interface {
	Create(ctx context.Context, tr entity.Transaction) error
	GetByUserID(ctx context.Context, userID int64) ([]entity.Transaction, error)
}

type MerchRepository interface {
	List(ctx context.Context) ([]entity.MerchItem, error)
	GetByID(ctx context.Context, id int64) (entity.MerchItem, error)
	GetByName(ctx context.Context, name string) (entity.MerchItem, error)
}

type InventoryRepository interface {
	GetByUserID(ctx context.Context, userID int64) ([]entity.UserInventory, error)
	Update(ctx context.Context, inventory entity.UserInventory) error
	Create(ctx context.Context, inventory entity.UserInventory) error
}

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
}
