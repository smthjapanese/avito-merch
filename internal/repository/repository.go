package repository

import (
	"context"
	"github.com/smthjapanese/avito-merch/internal/entity"
)

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
