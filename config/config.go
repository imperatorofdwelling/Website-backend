package config

import (
	"fmt"
	"log"

	"log/slog"

	"github.com/imperatorofdwelling/Website-backend/internal/metrics"
	"github.com/imperatorofdwelling/Website-backend/internal/server/http"

	"github.com/imperatorofdwelling/Website-backend/pkg/repository/postgres"
	"github.com/imperatorofdwelling/Website-backend/pkg/repository/redis"

	"github.com/joho/godotenv"
)

type Config struct {
	Server            *http.ServerConfig
	PostgresSQLConfig *postgres.PostgresSQLConfig
	RedisConfig       *redis.RedisConfig
}

func LoadConfig(envFilePath string) *Config {
	err := loadDotEnv(envFilePath)
	if err != nil {
		log.Fatal(err)
	}
	serverConfig, err := http.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	postgresSQLConfig, err := postgres.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	redisConfig, err := redis.LoadRedisConfig()
	if err != nil {
		log.Fatal(err)
	}
	cfg := &Config{
		Server:            serverConfig,
		PostgresSQLConfig: postgresSQLConfig,
		RedisConfig:       redisConfig,
	}
	fmt.Printf("%+v\n", cfg.RedisConfig)
	return cfg
}

func (c *Config) Run(logger *slog.Logger) {
	// PostgresSQL
	err := postgres.InitPostgresDB(c.PostgresSQLConfig)
	if err != nil {
		log.Fatal(err)
	}
	db, _ := postgres.GetDB()
	repo := postgres.NewLogRepository(db)

	// Redis
	//err = redis.InitRedis(c.RedisConfig)
	//if err != nil {
	//	log.Fatal(err)
	//}

	// To init storeId and secretKey from .env
	metrics.Init()
	srv := http.New(c.Server, logger, repo)

	defer c.Disconnect(srv)
	srv.Run()
}

func (c *Config) Disconnect(server *http.Server) {
	err := postgres.Disconnect()
	if err != nil {
		log.Println(err)
	}
	err = redis.Disconnect()
	if err != nil {
		log.Println(err)
	}
	err = server.Disconnect()
	if err != nil {
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
