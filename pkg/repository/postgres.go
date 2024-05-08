package repository

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"sync"
)

const (
	usersTable = "users"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

type PostgresDB struct {
	sync.Mutex
	db *sqlx.DB
}

var currDB *PostgresDB

// InitPostgresDB initializes a new instance of the Database.
func InitPostgresDB(cfg Config) error {
	// This if ensures that only 1 database instance is initialized.
	if currDB != nil {
		return errors.New("the database is already initialized")
	}
	sqlxDB, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode))
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
