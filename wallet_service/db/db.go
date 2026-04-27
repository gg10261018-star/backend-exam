package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init() error {
	var err error
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=wallet sslmode=disable"

	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	if err = migrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	fmt.Println("DB connected and migrated")
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func migrate() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id      BIGSERIAL PRIMARY KEY,
			name    VARCHAR(100) NOT NULL,
			balance NUMERIC(20, 2) NOT NULL DEFAULT 10000.00,
			CONSTRAINT balance_non_negative CHECK (balance >= 0)
		);
	`)
	return err
}
