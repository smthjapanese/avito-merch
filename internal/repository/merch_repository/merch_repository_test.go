package merch_repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/smthjapanese/avito-merch/internal/entity"
	"github.com/smthjapanese/avito-merch/internal/testutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type MerchRepositoryTestSuite struct {
	suite.Suite
	db   *sqlx.DB
	repo *MerchRepository
}

func (s *MerchRepositoryTestSuite) SetupSuite() {
	db, err := testutils.GetTestDB()
	require.NoError(s.T(), err)

	s.db = db
	s.repo = NewMerchRepository(db)

	s.recreateTable()
}

func (s *MerchRepositoryTestSuite) SetupTest() {
	_, err := s.db.Exec("TRUNCATE TABLE merch_items RESTART IDENTITY")
	require.NoError(s.T(), err)
}

func (s *MerchRepositoryTestSuite) TearDownSuite() {
	s.db.Close()
}

func (s *MerchRepositoryTestSuite) recreateTable() {
	_, err := s.db.Exec(`
        DROP TABLE IF EXISTS merch_items;
        CREATE TABLE merch_items (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) UNIQUE NOT NULL,
            price INTEGER NOT NULL CHECK (price > 0),
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        )
    `)
	require.NoError(s.T(), err)
}

func (s *MerchRepositoryTestSuite) TestList() {
	ctx := context.Background()

	// Insert test data
	testItems := []entity.MerchItem{
		{
			Name:      "Item1",
			Price:     100,
			CreatedAt: time.Now().UTC(),
		},
		{
			Name:      "Item2",
			Price:     200,
			CreatedAt: time.Now().UTC(),
		},
	}

	for _, item := range testItems {
		_, err := s.db.Exec(`
            INSERT INTO merch_items (name, price, created_at)
            VALUES ($1, $2, $3)
        `, item.Name, item.Price, item.CreatedAt)
		s.NoError(err)
	}

	// Test listing
	items, err := s.repo.List(ctx)
	s.NoError(err)
	s.Len(items, 2)
	s.Equal(testItems[0].Name, items[0].Name)
	s.Equal(testItems[0].Price, items[0].Price)
}

func (s *MerchRepositoryTestSuite) TestGetByID() {
	ctx := context.Background()

	// Insert test item
	testItem := entity.MerchItem{
		Name:      "TestItem",
		Price:     100,
		CreatedAt: time.Now().UTC(),
	}

	var itemID int64
	err := s.db.QueryRow(`
        INSERT INTO merch_items (name, price, created_at)
        VALUES ($1, $2, $3)
        RETURNING id
    `, testItem.Name, testItem.Price, testItem.CreatedAt).Scan(&itemID)
	s.NoError(err)

	s.Run("existing item", func() {
		found, err := s.repo.GetByID(ctx, itemID)
		s.NoError(err)
		s.Equal(testItem.Name, found.Name)
		s.Equal(testItem.Price, found.Price)
		s.Equal(itemID, found.ID)
	})

	s.Run("non-existing item", func() {
		_, err := s.repo.GetByID(ctx, 99999)
		s.ErrorIs(err, entity.ErrMerchNotFound)
	})
}

func (s *MerchRepositoryTestSuite) TestGetByName() {
	ctx := context.Background()

	// Insert test item
	testItem := entity.MerchItem{
		Name:      "TestItem",
		Price:     100,
		CreatedAt: time.Now().UTC(),
	}

	_, err := s.db.Exec(`
        INSERT INTO merch_items (name, price, created_at)
        VALUES ($1, $2, $3)
    `, testItem.Name, testItem.Price, testItem.CreatedAt)
	s.NoError(err)

	s.Run("existing item", func() {
		found, err := s.repo.GetByName(ctx, testItem.Name)
		s.NoError(err)
		s.Equal(testItem.Name, found.Name)
		s.Equal(testItem.Price, found.Price)
	})

	s.Run("non-existing item", func() {
		_, err := s.repo.GetByName(ctx, "NonExistentItem")
		s.ErrorIs(err, entity.ErrMerchNotFound)
	})

	s.Run("case sensitive search", func() {
		_, err := s.repo.GetByName(ctx, "testitem")
		s.ErrorIs(err, entity.ErrMerchNotFound)
	})
}

func (s *MerchRepositoryTestSuite) TestTransaction() {
	ctx := context.Background()
	s.Run("successful transaction", func() {
		tx, err := s.db.BeginTxx(ctx, nil)
		s.NoError(err)

		txRepo := s.repo.WithTx(tx)

		items, err := txRepo.List(ctx)
		s.NoError(err)
		initialCount := len(items)

		// Insert test data within transaction
		_, err = tx.Exec(`
            INSERT INTO merch_items (name, price, created_at)
            VALUES ($1, $2, $3)
        `, "TxItem", 100, time.Now().UTC())
		s.NoError(err)

		// Check item is visible within transaction
		items, err = txRepo.List(ctx)
		s.NoError(err)
		s.Equal(initialCount+1, len(items))

		// Check item is not visible outside transaction
		itemsOutside, err := s.repo.List(ctx)
		s.NoError(err)
		s.Equal(initialCount, len(itemsOutside))

		err = tx.Commit()
		s.NoError(err)

		// Verify item persists after commit
		items, err = s.repo.List(ctx)
		s.NoError(err)
		s.Equal(initialCount+1, len(items))
	})

	s.Run("rollback transaction", func() {
		tx, err := s.db.BeginTxx(ctx, nil)
		s.NoError(err)

		txRepo := s.repo.WithTx(tx)

		items, err := txRepo.List(ctx)
		s.NoError(err)
		initialCount := len(items)

		// Insert test data within transaction
		_, err = tx.Exec(`
            INSERT INTO merch_items (name, price, created_at)
            VALUES ($1, $2, $3)
        `, "RollbackItem", 100, time.Now().UTC())
		s.NoError(err)

		// Check item is visible within transaction
		items, err = txRepo.List(ctx)
		s.NoError(err)
		s.Equal(initialCount+1, len(items))

		err = tx.Rollback()
		s.NoError(err)

		// Verify item doesn't exist after rollback
		items, err = s.repo.List(ctx)
		s.NoError(err)
		s.Equal(initialCount, len(items))
	})

	s.Run("transaction isolation", func() {
		// Start first transaction
		tx1, err := s.db.BeginTxx(ctx, nil)
		s.NoError(err)
		txRepo1 := s.repo.WithTx(tx1)

		// Start second transaction
		tx2, err := s.db.BeginTxx(ctx, nil)
		s.NoError(err)
		txRepo2 := s.repo.WithTx(tx2)

		// Insert item in first transaction
		_, err = tx1.Exec(`
            INSERT INTO merch_items (name, price, created_at)
            VALUES ($1, $2, $3)
        `, "IsolationItem", 100, time.Now().UTC())
		s.NoError(err)

		// Check item is visible in first transaction
		items1, err := txRepo1.List(ctx)
		s.NoError(err)
		count1 := len(items1)

		// Check item is not visible in second transaction
		items2, err := txRepo2.List(ctx)
		s.NoError(err)
		s.Equal(count1-1, len(items2))

		// Commit first transaction
		err = tx1.Commit()
		s.NoError(err)

		// Now item should be visible in second transaction
		items2, err = txRepo2.List(ctx)
		s.NoError(err)
		s.Equal(count1, len(items2))

		err = tx2.Commit()
		s.NoError(err)
	})
}

func (s *MerchRepositoryTestSuite) TestListEmpty() {
	ctx := context.Background()

	items, err := s.repo.List(ctx)
	s.NoError(err)
	s.Empty(items)
}

func TestMerchRepository(t *testing.T) {
	suite.Run(t, new(MerchRepositoryTestSuite))
}
