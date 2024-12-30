package routes

import (
	"database/sql"

	"github.com/Rafin000/e-wallet/internal/common"
	"github.com/Rafin000/e-wallet/internal/domain"
	"github.com/Rafin000/e-wallet/internal/server/secure"
	"github.com/gin-gonic/gin"
)

func InitRoutes(rg *gin.RouterGroup, db *sql.DB, config *common.AppConfig, jm *secure.JWTManager, cardEncryptor *secure.CardEncryptor) {
	userRepo := domain.NewUserRepository(db)

	// Register public routes
	// registerAuthRou/home/raf1n/my_projects/xpay/internal/server/routes/user.gotes(rg, userRepo, jm)

	// Create authenticated user gin router group
	authGroup := rg.Group("/users")
	// authGroup.Use(middlewares.AuthMiddleware(userRepo, jm.GetPublicKey(), rbac))

	// Register authenticated routes
	registerUserManagementRoutes(authGroup, userRepo)
}
