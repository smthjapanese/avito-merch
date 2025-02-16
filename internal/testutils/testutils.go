package testutils

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
)

func GetTestDB() (*sqlx.DB, error) {
	host := GetEnv("TEST_DB_HOST", "localhost")
	port := GetEnv("TEST_DB_PORT", "5432")
	user := GetEnv("TEST_DB_USER", "avito")
	password := GetEnv("TEST_DB_PASSWORD", "avito_pass")
	dbname := GetEnv("TEST_DB_NAME", "test_db")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	return sqlx.Connect("postgres", dsn)
}

func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
