package server

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Router *gin.Engine
	httpServer *http.Server
	DB *sql.DB
	
}