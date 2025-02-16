package user_repository

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

type UserRepositoryTestSuite struct {
	suite.Suite
	db   *sqlx.DB
	repo *UserRepository
}

func (s *UserRepositoryTestSuite) SetupSuite() {
	// Получаем параметры подключения из переменных окружения или используем значения по умолчанию
	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5432")
	user := getEnv("TEST_DB_USER", "postgres")
	password := getEnv("TEST_DB_PASSWORD", "postgres")
	dbname := getEnv("TEST_DB_NAME", "test_db")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(s.T(), err)

	s.db = db
	s.repo = NewUserRepository(db)

	// Создаем таблицу
	s.recreateTable()
}

func (s *UserRepositoryTestSuite) SetupTest() {
	// Очищаем таблицу перед каждым тестом
	_, err := s.db.Exec("TRUNCATE TABLE users RESTART IDENTITY")
	require.NoError(s.T(), err)
}

func (s *UserRepositoryTestSuite) TearDownSuite() {
	s.db.Close()
}

func (s *UserRepositoryTestSuite) recreateTable() {
	_, err := s.db.Exec(`
  DROP TABLE IF EXISTS users;
  CREATE TABLE users (
   id SERIAL PRIMARY KEY,
   username VARCHAR(255) UNIQUE NOT NULL,
   password_hash VARCHAR(255) NOT NULL,
   coins INTEGER NOT NULL DEFAULT 1000,
   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
  )
 `)
	require.NoError(s.T(), err)
}

func (s *UserRepositoryTestSuite) TestCreateUser() {
	ctx := context.Background()

	s.Run("successful creation", func() {
		user := &entity.User{
			Username:     "testuser",
			PasswordHash: "hashed_password",
			Coins:        1000,
			CreatedAt:    time.Now().UTC(),
		}

		err := s.repo.Create(ctx, user)
		s.NoError(err)
		s.NotZero(user.ID)
	})

	s.Run("duplicate username", func() {
		user := &entity.User{
			Username:     "testuser",
			PasswordHash: "another_password",
			Coins:        1000,
			CreatedAt:    time.Now().UTC(),
		}

		err := s.repo.Create(ctx, user)
		s.ErrorIs(err, entity.ErrUserAlreadyExists)
	})
}

func (s *UserRepositoryTestSuite) TestGetByID() {
	ctx := context.Background()

	// Создаем тестового пользователя
	user := &entity.User{
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Coins:        1000,
		CreatedAt:    time.Now().UTC(),
	}
	err := s.repo.Create(ctx, user)
	s.NoError(err)

	s.Run("existing user", func() {
		found, err := s.repo.GetByID(ctx, user.ID)
		s.NoError(err)
		s.Equal(user.Username, found.Username)
		s.Equal(user.PasswordHash, found.PasswordHash)
		s.Equal(user.Coins, found.Coins)
	})

	s.Run("non-existing user", func() {
		found, err := s.repo.GetByID(ctx, 999999)
		s.ErrorIs(err, entity.ErrUserNotFound)
		s.Nil(found)
	})
}

func (s *UserRepositoryTestSuite) TestTransaction() {
	ctx := context.Background()

	s.Run("successful transaction", func() {
		tx, err := s.db.BeginTxx(ctx, nil)
		s.NoError(err)

		txRepo := s.repo.WithTx(tx)

		user := &entity.User{
			Username:     "txuser",
			PasswordHash: "hashed_password",
			Coins:        1000,
			CreatedAt:    time.Now().UTC(),
		}

		err = txRepo.Create(ctx, user)
		s.NoError(err)

		err = tx.Commit()
		s.NoError(err)

		found, err := s.repo.GetByUsername(ctx, "txuser")
		s.NoError(err)
		s.Equal(user.Username, found.Username)
	})

	s.Run("rollback transaction", func() {
		tx, err := s.db.BeginTxx(ctx, nil)
		s.NoError(err)

		txRepo := s.repo.WithTx(tx)

		user := &entity.User{
			Username:     "rollbackuser",
			PasswordHash: "hashed_password",
			Coins:        1000,
			CreatedAt:    time.Now().UTC(),
		}

		err = txRepo.Create(ctx, user)
		s.NoError(err)

		err = tx.Rollback()
		s.NoError(err)

		found, err := s.repo.GetByUsername(ctx, "rollbackuser")
		s.ErrorIs(err, entity.ErrUserNotFound)
		s.Nil(found)
	})
}

func TestUserRepository(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
