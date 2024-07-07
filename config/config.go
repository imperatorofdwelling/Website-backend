package config

import (
	"github.com/https-whoyan/dwellingPayload/internal/metrics"
	"log"
	"log/slog"

	"github.com/joho/godotenv"

	"github.com/https-whoyan/dwellingPayload/internal/server/http"
	"github.com/https-whoyan/dwellingPayload/pkg/repository"
)

type Config struct {
	Server            *http.ServerConfig
	PostgresSQLConfig *repository.PostgresSQLConfig
}

func LoadConfig(envFilePath string) *Config {
	err := loadDotEnv(envFilePath)
	if err != nil {
		//log Fatal by logger
		log.Fatal(err)
	}
	serverConfig, err := http.LoadConfig()
	if err != nil {
		//log Fatal by logger
		log.Fatal(err)
	}
	postgresSQLConfig, err := repository.LoadConfig()
	if err != nil {
		//log Fatal by logger
		log.Fatal(err)
	}
	cfg := &Config{
		Server:            serverConfig,
		PostgresSQLConfig: postgresSQLConfig,
	}
	return cfg
}

func (c *Config) Run(logger *slog.Logger) {
	err := repository.InitPostgresDB(c.PostgresSQLConfig)
	if err != nil {
		//log Fatal by logger
		log.Fatal(err)
	}
	db, _ := repository.GetDB()
	repo := repository.NewLogRepository(db)
	// To init storeId and secretKey from .env
	metrics.Init()
	srv := http.New(c.Server, logger, repo)

	defer c.Disconnect(srv)
	srv.Run()
}

func (c *Config) Disconnect(server *http.Server) {
	err := repository.Disconnect()
	if err != nil {
		// Log print by logger
		log.Println(err)
	}
	err = server.Disconnect()
	if err != nil {
		// Log print by logger
		log.Println(err)
	}
}

func loadDotEnv(filePath string) error {
	if filePath == "" {
		filePath = ".env"
	}
	err := godotenv.Load(filePath)
	return err
}
