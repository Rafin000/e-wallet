package routes

import (
	"github.com/Rafin000/e-wallet/internal/domain"
	"github.com/gin-gonic/gin"
)

func registerUserManagementRoutes(rg *gin.RouterGroup, userRepo domain.UserRepository) {
	userHandler := handlers.NewUserHandler(userRepo)
	rg.POST("", userHandler.CreateUserWithRole)
}
