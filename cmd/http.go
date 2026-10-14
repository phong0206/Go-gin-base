package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phong0206/go_gin_restful/internal/common/helper"
	"github.com/phong0206/go_gin_restful/internal/config"
	"github.com/phong0206/go_gin_restful/internal/domain/repositories"
	"github.com/phong0206/go_gin_restful/internal/handler"
	"github.com/phong0206/go_gin_restful/internal/repositories/postgres"
	"github.com/phong0206/go_gin_restful/internal/router"
	"github.com/phong0206/go_gin_restful/internal/usecases"
	"github.com/phong0206/go_gin_restful/pkg/database"
	"github.com/phong0206/go_gin_restful/pkg/jwt"
)

type HTTPServer struct {
	server *http.Server
}

func NewHTTPServer(router *gin.Engine, addr string) *HTTPServer {
	return &HTTPServer{
		server: &http.Server{
			Addr:              addr,
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

func (s *HTTPServer) Start() error {
	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func runHTTP() error {
	// 1. Load application configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// 2. Initialize database connection pool
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Database.ConnectTimeout)
	db, err := database.OpenDatabase(ctx, cfg.Database)
	cancel()
	if err != nil {
		log.Printf("[WARN] Database connection failed: %v (continuing in degraded mode)", err)
	} else {
		defer func() {
			if closeErr := db.Close(); closeErr != nil {
				log.Printf("[ERROR] Failed to close database: %v", closeErr)
			}
		}()
		log.Println("[INFO] Successfully connected to database")
	}

	// 3. Initialize layers (Repository -> Usecase -> Handler)
	var userRepo repositories.IUserRepository
	if db != nil {
		userRepo = postgres.NewUserRepository(db)
	}
	jwtAdapter := jwt.NewJWTService(cfg.JWT) 
	hasher := helper.NewPasswordHelper()

	authUsecase := usecases.NewAuthUsecase(userRepo, jwtAdapter, hasher)
	authHandler := handler.NewAuthHandler(authUsecase)

	userUsecase := usecases.NewUserUsecase(userRepo)
	userHandler := handler.NewUserHandler(userUsecase)

	// 4. Setup Gin router and routes
	r := router.NewRouter(cfg, userHandler, authHandler)

	// 5. Create and start HTTP Server
	server := NewHTTPServer(r, cfg.HTTP.Address())
	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("[INFO] %s is listening on %s (env: %s)", cfg.App.Name, cfg.HTTP.Address(), cfg.App.Environment)
		serverErrors <- server.Start()
	}()

	// 5. Graceful shutdown on OS interrupt
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		log.Printf("[INFO] Signal received: %v. Initiating graceful shutdown...", sig)
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}
		log.Println("[INFO] Server exited cleanly")
	}

	return nil
}
