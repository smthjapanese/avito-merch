package transaction_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/smthjapanese/avito-merch/internal/entity"
)

type Repository interface {
	Create(ctx context.Context, tr *entity.Transaction) error
	GetByUserID(ctx context.Context, userID int64) ([]entity.Transaction, error)
}

type dbConn interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type TransactionRepository struct {
	db dbConn
}

func NewTransactionRepository(db *sqlx.DB) *TransactionRepository {
	return &TransactionRepository{
		db: db,
	}
}

func (r *TransactionRepository) WithTx(tx *sqlx.Tx) *TransactionRepository {
	return &TransactionRepository{
		db: tx,
	}
}

func (r *TransactionRepository) Create(ctx context.Context, tr *entity.Transaction) error {
	query := `
  INSERT INTO transactions (from_user_id, to_user_id, amount, type, item_id)
  VALUES ($1, $2, $3, $4, $5)
  RETURNING id, created_at`

	err := r.db.QueryRowContext(
		ctx,
		query,
		tr.FromUserID,
		tr.ToUserID,
		tr.Amount,
		tr.Type,
		tr.ItemID,
	).Scan(&tr.ID, &tr.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (r *TransactionRepository) GetByUserID(ctx context.Context, userID int64) ([]entity.Transaction, error) {
	query := `
  SELECT id, from_user_id, to_user_id, amount, type, item_id, created_at
  FROM transactions
  WHERE from_user_id = $1 OR to_user_id = $1
  ORDER BY created_at DESC`

	var transactions []entity.Transaction
	err := r.db.SelectContext(ctx, &transactions, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []entity.Transaction{}, nil
		}
		return nil, fmt.Errorf("failed to get transactions by user id: %w", err)
	}

	return transactions, nil
}
