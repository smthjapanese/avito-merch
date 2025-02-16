package transaction_repository

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/smthjapanese/avito-merch/internal/entity"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
)

type TransactionRepositoryTestSuite struct {
	suite.Suite
	db   *sqlx.DB
	repo *TransactionRepository
}

func (s *TransactionRepositoryTestSuite) SetupSuite() {
	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5432")
	user := getEnv("TEST_DB_USER", "avito")              // Замените на ваш username
	password := getEnv("TEST_DB_PASSWORD", "avito_pass") // Замените на ваш пароль
	dbname := getEnv("TEST_DB_NAME", "test_db")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(s.T(), err)

	s.db = db
	s.repo = NewTransactionRepository(db)

	s.recreateTables()
}

func (s *TransactionRepositoryTestSuite) SetupTest() {
	_, err := s.db.Exec("TRUNCATE TABLE transactions, users RESTART IDENTITY CASCADE")
	require.NoError(s.T(), err)

	_, err = s.db.Exec(`
  INSERT INTO users (username, password_hash, coins) 
  VALUES 
  ('user1', 'hash1', 1000),
  ('user2', 'hash2', 1000)`)
	require.NoError(s.T(), err)
}

func (s *TransactionRepositoryTestSuite) TearDownSuite() {
	s.db.Close()
}

func (s *TransactionRepositoryTestSuite) recreateTables() {
	_, err := s.db.Exec(`
  DROP TABLE IF EXISTS transactions;
  DROP TABLE IF EXISTS users;
  
  CREATE TABLE users (
   id SERIAL PRIMARY KEY,
   username VARCHAR(255) UNIQUE NOT NULL,
   password_hash VARCHAR(255) NOT NULL,
   coins INTEGER NOT NULL DEFAULT 1000,
   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
  );
  
  CREATE TABLE transactions (
   id SERIAL PRIMARY KEY,
   from_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
   to_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
   amount INTEGER NOT NULL CHECK (amount > 0),
   type VARCHAR(50) NOT NULL CHECK (type IN ('transfer', 'purchase')),
   item_id INTEGER,
   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
  );
  
  CREATE INDEX idx_transactions_from_user ON transactions(from_user_id);
  CREATE INDEX idx_transactions_to_user ON transactions(to_user_id);
  CREATE INDEX idx_transactions_created_at ON transactions(created_at);
 `)
	require.NoError(s.T(), err)
}

func (s *TransactionRepositoryTestSuite) TestCreateTransaction() {
	ctx := context.Background()

	s.Run("successful transfer creation", func() {
		tx := entity.Transaction{
			FromUserID: 1,
			ToUserID:   2,
			Amount:     100,
			Type:       entity.TransactionTypeTransfer,
		}

		err := s.repo.Create(ctx, &tx)
		s.NoError(err)
		s.NotZero(tx.ID)
		s.NotZero(tx.CreatedAt)
	})

	s.Run("successful purchase creation", func() {
		itemID := int64(1)
		tx := entity.Transaction{
			FromUserID: 1,
			ToUserID:   2,
			Amount:     100,
			Type:       entity.TransactionTypePurchase,
			ItemID:     &itemID,
		}

		err := s.repo.Create(ctx, &tx)
		s.NoError(err)
		s.NotZero(tx.ID)
		s.NotZero(tx.CreatedAt)
	})

	s.Run("fail on non-existent user", func() {
		tx := entity.Transaction{
			FromUserID: 999,
			ToUserID:   1,
			Amount:     100,
			Type:       entity.TransactionTypeTransfer,
		}

		err := s.repo.Create(ctx, &tx)
		s.Error(err)
	})
}

func (s *TransactionRepositoryTestSuite) TestGetByUserID() {
	ctx := context.Background()

	// Create test transactions
	txs := []entity.Transaction{
		{
			FromUserID: 1,
			ToUserID:   2,
			Amount:     100,
			Type:       entity.TransactionTypeTransfer,
		},
		{
			FromUserID: 2,
			ToUserID:   1,
			Amount:     50,
			Type:       entity.TransactionTypeTransfer,
		},
	}

	for _, tx := range txs {
		err := s.repo.Create(ctx, &tx)
		s.NoError(err)
		time.Sleep(time.Millisecond)
	}
	s.Run("get user1 transactions", func() {
		transactions, err := s.repo.GetByUserID(ctx, 1)
		s.NoError(err)
		s.Len(transactions, 2)
		s.True(transactions[0].CreatedAt.After(transactions[1].CreatedAt))
	})

	s.Run("get user2 transactions", func() {
		transactions, err := s.repo.GetByUserID(ctx, 2)
		s.NoError(err)
		s.Len(transactions, 2)
	})

	s.Run("get non-existent user transactions", func() {
		transactions, err := s.repo.GetByUserID(ctx, 999)
		s.NoError(err)
		s.Len(transactions, 0)
	})
}

func (s *TransactionRepositoryTestSuite) TestTransaction() {
	ctx := context.Background()

	s.Run("successful transaction", func() {
		tx, err := s.db.BeginTxx(ctx, nil)
		s.NoError(err)

		txRepo := s.repo.WithTx(tx)

		transaction := entity.Transaction{
			FromUserID: 1,
			ToUserID:   2,
			Amount:     100,
			Type:       entity.TransactionTypeTransfer,
		}

		err = txRepo.Create(ctx, &transaction)
		s.NoError(err)

		err = tx.Commit()
		s.NoError(err)

		transactions, err := s.repo.GetByUserID(ctx, 1)
		s.NoError(err)
		s.Len(transactions, 1)
	})

	s.Run("rollback transaction", func() {
		// Очищаем таблицу перед тестом
		_, err := s.db.Exec("TRUNCATE TABLE transactions RESTART IDENTITY")
		s.NoError(err)

		tx, err := s.db.BeginTxx(ctx, nil)
		s.NoError(err)

		txRepo := s.repo.WithTx(tx)

		transaction := entity.Transaction{
			FromUserID: 1,
			ToUserID:   2,
			Amount:     100,
			Type:       entity.TransactionTypeTransfer,
		}

		err = txRepo.Create(ctx, &transaction)
		s.NoError(err)

		var countBefore int
		err = tx.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&countBefore)
		s.NoError(err)
		s.Equal(1, countBefore, "Should have one transaction before rollback")

		err = tx.Rollback()
		s.NoError(err)

		var countAfter int
		err = s.db.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&countAfter)
		s.NoError(err)
		s.Equal(0, countAfter, "Should have no transactions after rollback")
	})
}

func TestTransactionRepository(t *testing.T) {
	suite.Run(t, new(TransactionRepositoryTestSuite))
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
