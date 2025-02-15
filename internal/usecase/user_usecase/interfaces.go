package usecase

import (
	"context"
	"github.com/smthjapanese/avito-merch/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks/mocks.go -package=mocks

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
