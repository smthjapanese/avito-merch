package user_usecaseold

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/smthjapanese/avito-merch/internal/entity"
)

type UserUseCase struct {
	userRepo UserRepository
	clock    Clock
}

func NewUserUseCase(userRepo UserRepository, clock Clock) UserUseCase {
	return UserUseCase{
		userRepo: userRepo,
		clock:    clock,
	}
}

func (uc *UserUseCase) Create(ctx context.Context, username, password string) (entity.User, error) {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	passwordHash := hex.EncodeToString(hasher.Sum(nil))

	user := entity.User{
		Username:     username,
		PasswordHash: passwordHash,
		Coins:        entity.InitialBalance,
		CreatedAt:    uc.clock.Now(),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (uc *UserUseCase) GetByID(ctx context.Context, id int64) (entity.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}

func (uc *UserUseCase) GetByUsername(ctx context.Context, username string) (entity.User, error) {
	user, err := uc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return entity.User{}, entity.ErrUserNotFound
	}
	return user, nil
}

func (uc *UserUseCase) UpdateUser(ctx context.Context, user entity.User) error {
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return err
	}

	return nil
}
