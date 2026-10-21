package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/phong0206/go_gin_restful/internal/config"
	"github.com/phong0206/go_gin_restful/internal/handler"
	"github.com/phong0206/go_gin_restful/internal/middleware"
)

// NewRouter initializes and configures the Gin engine with all global middlewares and routes.
func NewRouter(cfg *config.Config, userHandler *handler.UserHandler, authHandler *handler.AuthHandler, verifyToken gin.HandlerFunc) *gin.Engine {
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()

	// Global middlewares
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.RequestID())
	r.Use(middleware.AccessLogger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins, cfg.CORS.AllowCredentials))
	r.Use(middleware.SecurityHeaders(cfg.App.Environment == "production"))
	r.Use(middleware.ResponseInterceptor())

	// Custom 404 & 405 handlers
	r.NoRoute(middleware.NotFound)
	r.NoMethod(middleware.MethodNotAllowed)

	// System health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"app":    cfg.App.Name,
			"env":    cfg.App.Environment,
		})
	})

	// API version group (e.g., /api/v1)
	api := r.Group(cfg.HTTP.APIPrefix)
	{
		RegisterAuthRoutes(api, authHandler)
		RegisterUserRoutes(api, userHandler, verifyToken)
	}

	return r
}
