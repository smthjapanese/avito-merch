// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
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
	GetInfo(ctx context.Context, userID int64) (*entity.InfoResponse, error)
	SendCoins(ctx context.Context, fromUserID int64, toUser string, amount int64) error
	BuyMerch(ctx context.Context, userID int64, merchName string) error
}

type Repository interface {
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
