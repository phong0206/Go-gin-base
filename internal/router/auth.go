package router

import (
	"github.com/gin-gonic/gin"
	"github.com/phong0206/go_gin_restful/internal/handler"
)

// RegisterAuthRoutes sets up auth-related endpoints under the provided router group.
func RegisterAuthRoutes(rg *gin.RouterGroup, authHandler *handler.AuthHandler) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
	}
}
