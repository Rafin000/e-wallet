package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Rafin000/e-wallet/internal/common"
	"github.com/Rafin000/e-wallet/internal/infra/postgres"
	"github.com/Rafin000/e-wallet/internal/server/middlewares"
	"github.com/Rafin000/e-wallet/internal/server/routes"
	"github.com/Rafin000/e-wallet/internal/server/secure"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	db, err := setupPostgres(ctx, cfg.DB)
	if err != nil {
		return nil, err
	}

	router := setupRouter(cfg.App)

	jwtManager, err := secure.NewJWTManager(&cfg.JWT)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT manager: %w", err)
	}

	cardEncryptor, err := secure.NewCardEncryptor(cfg.Card.AESKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create card encryptor: %w", err)
	}

	s := &Server{
		Router: router,
		DB:     db,
		Config: cfg,
		httpServer: &http.Server{
			Addr:         cfg.App.ServerAddress,
			Handler:      router,
			IdleTimeout:  common.Timeouts.Server.Read * 2,
			ReadTimeout:  common.Timeouts.Server.Read,
			WriteTimeout: common.Timeouts.Server.Write,
		},
	}
	s.setupRoutes(jwtManager, cardEncryptor)
	s.setupMiddlewares()
	// setSwaggerInfo(s.httpServer.Addr)

	return s, nil
}

// Start begins listening for HTTP requests on the configured address.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// setupRoutes initializes all API routes for the server.
func (s *Server) setupRoutes(jm *secure.JWTManager, cardEncryptor *secure.CardEncryptor) {
	s.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiGroup := s.Router.Group("/api/v1")
	routes.InitRoutes(apiGroup, s.DB, s.Config, jm, cardEncryptor)
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

// setupRouter initializes and configures the Gin router.
// It sets the Gin mode based on the application settings and disables trusted proxies.
func setupRouter(appSettings common.AppSettings) *gin.Engine {
	gin.SetMode(appSettings.GinMode)
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	return router
}

// setupMiddlewares adds all necessary middlewares to the Gin router.
func (s *Server) setupMiddlewares() {
	s.Router.Use(middlewares.InitMiddlewares()...)
}

// setSwaggerInfo configures Swagger documentation settings for the API.
// func setSwaggerInfo(addr string) {
// 	docs.SwaggerInfo.Title = "xPay Digital Wallet API"
// 	docs.SwaggerInfo.Version = "1.0"
// 	docs.SwaggerInfo.Schemes = []string{"http", "https"}
// 	docs.SwaggerInfo.BasePath = "/api/v1"
// 	docs.SwaggerInfo.Host = addr

// 	slog.Info(fmt.Sprintf("Swagger Specs available at http://%s/swagger/index.html", docs.SwaggerInfo.Host))
// }

// Shutdown gracefully stops the server, closing the database connection and stopping the HTTP server.
// It uses the provided context for timeout control.
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.DB.Close(); err != nil {
		slog.Error("failed to close database connection", "error", err)
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}
