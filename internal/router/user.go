package router

import (
	"github.com/gin-gonic/gin"
	"github.com/phong0206/go_gin_restful/internal/handler"
)

// RegisterUserRoutes sets up user-related endpoints under the provided router group.
func RegisterUserRoutes(rg *gin.RouterGroup, userHandler *handler.UserHandler) {
	users := rg.Group("/users")
	{
		users.GET("", userHandler.List)
		users.GET("/:id", userHandler.GetByID)
	}
}
