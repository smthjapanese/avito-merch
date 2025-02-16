package user_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/smthjapanese/avito-merch/internal/entity"
)

type dbConn interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type UserRepository struct {
	db dbConn
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) WithTx(tx *sqlx.Tx) *UserRepository {
	return &UserRepository{
		db: tx,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
  INSERT INTO users (username, password_hash, coins, created_at)
  VALUES ($1, $2, $3, $4)
  RETURNING id`

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.PasswordHash,
		user.Coins,
		user.CreatedAt,
	).Scan(&user.ID)

	if err != nil {
		if isUniqueViolation(err) {
			return entity.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	var user entity.User
	query := `
  SELECT id, username, password_hash, coins, created_at
  FROM users
  WHERE id = $1`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	query := `
  SELECT id, username, password_hash, coins, created_at
  FROM users
  WHERE username = $1`

	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
  UPDATE users
  SET username = $1,
   password_hash = $2,
   coins = $3
  WHERE id = $4`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.PasswordHash,
		user.Coins,
		user.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return entity.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entity.ErrUserNotFound
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}
