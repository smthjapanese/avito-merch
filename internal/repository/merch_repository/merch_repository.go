package merch_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/smthjapanese/avito-merch/internal/entity"
)

// dbConn is an interface that both *sqlx.DB and *sqlx.Tx satisfy
type dbConn interface {
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type MerchRepository struct {
	db dbConn
}

func NewMerchRepository(db *sqlx.DB) *MerchRepository {
	return &MerchRepository{
		db: db,
	}
}

// WithTx returns a new MerchRepository that uses the transaction
func (r *MerchRepository) WithTx(tx *sqlx.Tx) *MerchRepository {
	return &MerchRepository{
		db: tx,
	}
}

// List returns all available merch items
func (r *MerchRepository) List(ctx context.Context) ([]entity.MerchItem, error) {
	var items []entity.MerchItem
	query := `
        SELECT id, name, price, created_at
        FROM merch_items
        ORDER BY id`

	err := r.db.SelectContext(ctx, &items, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list merch items: %w", err)
	}

	return items, nil
}

// GetByID retrieves a merch item by its ID
func (r *MerchRepository) GetByID(ctx context.Context, id int64) (entity.MerchItem, error) {
	var item entity.MerchItem
	query := `
        SELECT id, name, price, created_at
        FROM merch_items
        WHERE id = $1`

	err := r.db.GetContext(ctx, &item, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.MerchItem{}, entity.ErrMerchNotFound
		}
		return entity.MerchItem{}, fmt.Errorf("failed to get merch item by id: %w", err)
	}

	return item, nil
}

// GetByName retrieves a merch item by its name
func (r *MerchRepository) GetByName(ctx context.Context, name string) (entity.MerchItem, error) {
	var item entity.MerchItem
	query := `
        SELECT id, name, price, created_at
        FROM merch_items
        WHERE name = $1`

	err := r.db.GetContext(ctx, &item, query, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.MerchItem{}, entity.ErrMerchNotFound
		}
		return entity.MerchItem{}, fmt.Errorf("failed to get merch item by name: %w", err)
	}

	return item, nil
}
