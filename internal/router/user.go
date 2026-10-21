package router

import (
	"github.com/gin-gonic/gin"
	"github.com/phong0206/go_gin_restful/internal/handler"
)

// RegisterUserRoutes sets up user-related endpoints under the provided router group.
func RegisterUserRoutes(rg *gin.RouterGroup, userHandler *handler.UserHandler, verifyToken gin.HandlerFunc) {
	users := rg.Group("/users")
	users.Use(verifyToken)
	{
		users.GET("", userHandler.List)
		users.GET("/:id", userHandler.GetByID)
	}
}
