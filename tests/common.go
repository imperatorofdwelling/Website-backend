package tests

import (
	"github.com/imperatorofdwelling/Website-backend/config"
	"log/slog"
	"net/http"

	srv "github.com/imperatorofdwelling/Website-backend/internal/server/http"
	internalLogger "github.com/imperatorofdwelling/Website-backend/pkg/logger"
	"github.com/imperatorofdwelling/Website-backend/pkg/repository/postgres"
)

var (
	dbCfg  = config.LoadConfig("../.env").PostgresSQLConfig
	logger = internalLogger.New(internalLogger.EnvLocal)
	router http.Handler
)

func Init() {
	if err := postgres.InitPostgresDB(dbCfg); err != nil {
		logger.Error("failed to init DB instance", slog.String("error", err.Error()))
	}
	db, _ := postgres.GetDB()
	logRepo := postgres.NewLogRepository(db)
	router = srv.NewRouter(logger, logRepo)
}
