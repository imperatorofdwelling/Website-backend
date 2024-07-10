package main

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/imperatorofdwelling/Website-backend/config"
	"log"
)

const (
	migrationsPath = "./pkg/repository/migrations"
)

func main() {
	cfg := config.LoadConfig("")
	dbCfg := cfg.PostgresSQLConfig

	url := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbCfg.DBUsername,
		dbCfg.DBPassword,
		dbCfg.DBHost,
		dbCfg.DBPort,
		dbCfg.DBName,
		dbCfg.DBSSLMode,
	)
	m, err := migrate.New(
		"file://"+migrationsPath,
		url,
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		log.Fatal(err)
	}
	fmt.Println("migrations applied successfully")
}
