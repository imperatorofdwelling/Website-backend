package tests

import (
	"log/slog"
	"net/http"

	"github.com/imperatorofdwelling/Website-backend/config"
	srv "github.com/imperatorofdwelling/Website-backend/internal/server/http"
	internalLogger "github.com/imperatorofdwelling/Website-backend/pkg/logger"
	"github.com/imperatorofdwelling/Website-backend/pkg/repository/postgres"
	"github.com/imperatorofdwelling/Website-backend/pkg/repository/redis"
)

var (
	cfg    = config.LoadConfig("../.env")
	logger = internalLogger.New(internalLogger.EnvLocal)
	router http.Handler
)

func Init() {
	// Initialize PostgreSQL
	if err := postgres.InitPostgresDB(cfg.PostgresSQLConfig); err != nil {
		logger.Error("failed to init DB instance", slog.String("error", err.Error()))
	}
	db, _ := postgres.GetDB()
	logRepo := postgres.NewLogRepository(db)

	// Initialize Redis
	if err := redis.InitRedis(cfg.RedisConfig); err != nil {
		logger.Error("failed to init Redis instance", slog.String("error", err.Error()))
	}
	redisDB, ok := redis.GetCurrRedisDB()
	if !ok {
		logger.Error("failed to get Redis connection")
		return
	}

	// Create router
	router = srv.NewRouter(logger, logRepo, redisDB)
}
