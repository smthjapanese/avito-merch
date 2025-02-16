package transaction_usecase

import (
	"context"
	"github.com/smthjapanese/avito-merch/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks/mocks.go -package=mocks

type Repository interface {
	Create(ctx context.Context, tr entity.Transaction) error
	GetByUserID(ctx context.Context, userID int64) ([]entity.Transaction, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
}

type DBTransactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
