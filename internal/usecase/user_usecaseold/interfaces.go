package user_usecaseold

import (
	"context"
	"github.com/smthjapanese/avito-merch/internal/entity"
	"time"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks/mocks.go -package=mocks

type UserRepository interface {
	Create(ctx context.Context, user entity.User) error
	GetByID(ctx context.Context, id int64) (entity.User, error)
	GetByUsername(ctx context.Context, username string) (entity.User, error)
	Update(ctx context.Context, user entity.User) error
}

type Clock interface {
	Now() time.Time
}
