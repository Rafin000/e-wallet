package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Rafin000/e-wallet/internal/common"
	"github.com/gin-gonic/gin"
)

// Server encapsulates all dependencies and configurations for the HTTP server.
type Server struct {
	Router     *gin.Engine
	httpServer *http.Server
	DB         *sql.DB
	Config     *common.AppConfig
}

func NewServer(ctx context.Context) (*Server, error) {
	cfg, err := common.LoadConfig()

	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	setupSlogger(cfg.App)

}

// Start begins listening for HTTP requests on the configured address.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// setupSlogger configures the global logger based on the application environment.
// It uses a text handler for development and a JSON handler for other environments.
func setupSlogger(appSettings common.AppSettings) {
	var logLevel = new(slog.LevelVar)
	var handler slog.Handler

	if appSettings.Env == common.AppEnvDev {
		handler = slog.NewTextHandler(os.Stderr, common.GetTextHandlerOptions(logLevel))
	} else {
		handler = slog.NewJSONHandler(os.Stderr, common.GetJSONHandlerOptions(logLevel))
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	if appSettings.GinMode == gin.DebugMode {
		logLevel.Set(slog.LevelDebug)
	}
}

// setupPostgres establishes a connection to the PostgreSQL database and runs migrations.
// It returns a database connection pool (*sql.DB) on success.
func setupPostgres(ctx context.Context, dbConfig common.DBConfig) (*sql.DB, error) {
	db, err := postgres.NewConnection(ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := postgres.RunMigrations(ctx, db); err != nil {
		slog.Warn("failed to run migrations", "err", err)
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}
