package routes

import (
	"github.com/Rafin000/e-wallet/internal/domain"
	"github.com/Rafin000/e-wallet/internal/secure"
	"github.com/Rafin000/e-wallet/internal/server/handlers"
	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(rg *gin.RouterGroup, userRepo domain.UserRepository, jm *secure.JWTManager) {
	authHandler := handlers.NewAuthHandler(userRepo, jm)

	rg.POST("/register", authHandler.Register)
	rg.POST("/login", authHandler.Login)
}
