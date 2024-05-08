package repository

import (
	"errors"
	"fmt"
	"github.com/https-whoyan/dwellingPayload/config"
	"github.com/jmoiron/sqlx"
	"sync"
)

const (
	usersTable = "users"
)

type PostgresDB struct {
	sync.Mutex
	db *sqlx.DB
}

var currDB *PostgresDB

// InitPostgresDB initializes a new instance of the Database.
func InitPostgresDB(cfg config.DB) error {
	// This if ensures that only 1 database instance is initialized.
	if currDB != nil {
		return errors.New("the database is already initialized")
	}
	sqlxDB, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUsername, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode))
	if err != nil {
		return err
	}

	if err = sqlxDB.Ping(); err != nil {
		return err
	}

	currDB = &PostgresDB{
		db: sqlxDB,
	}
	return nil
}

func GetDB() (*PostgresDB, bool) {
	if currDB == nil {
		return nil, false
	}
	return currDB, true
}
