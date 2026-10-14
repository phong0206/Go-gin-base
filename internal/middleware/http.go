package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	domainerrors "github.com/phong0206/go_gin_restful/internal/domain/errors"
	"github.com/phong0206/go_gin_restful/internal/common/constants"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(constants.HeaderRequestID)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set("request_id", requestID)
		c.Header(constants.HeaderRequestID, requestID)
		c.Next()
	}
}

func AccessLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		requestID, _ := c.Get("request_id")
		slog.Info("http request",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", time.Since(startedAt).String(),
			"client_ip", c.ClientIP(),
		)
	}
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		requestID, _ := c.Get("request_id")
		slog.Error("panic recovered", "request_id", requestID, "error", recovered)
		RespondError(c, domainerrors.New(domainerrors.KindInternal, "INTERNAL_SERVER_ERROR", "internal server error"))
	})
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		RespondError(c, c.Errors.Last().Err)
	}
}

func RespondError(c *gin.Context, err error) {
	appErr := domainerrors.Wrap(
		domainerrors.KindInternal,
		"INTERNAL_SERVER_ERROR",
		"internal server error",
		err,
	)
	var candidate *domainerrors.AppError
	if errors.As(err, &candidate) {
		appErr = candidate
	}

	if appErr.Kind == domainerrors.KindInternal {
		slog.Error("request error", "request_id", c.GetString("request_id"), "error", err)
	}

	c.AbortWithStatusJSON(statusFor(appErr.Kind), gin.H{
		"error": gin.H{
			"code":    appErr.Code,
			"message": appErr.Message,
		},
	})
}

func statusFor(kind domainerrors.Kind) int {
	switch kind {
	case domainerrors.KindValidation:
		return http.StatusBadRequest
	case domainerrors.KindUnauthorized:
		return http.StatusUnauthorized
	case domainerrors.KindForbidden:
		return http.StatusForbidden
	case domainerrors.KindNotFound:
		return http.StatusNotFound
	case domainerrors.KindConflict:
		return http.StatusConflict
	case domainerrors.KindMethod:
		return http.StatusMethodNotAllowed
	default:
		return http.StatusInternalServerError
	}
}

func SecurityHeaders(production bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
		if production {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

func CORS(allowedOrigins []string, allowCredentials bool) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", constants.HeaderRequestID},
		ExposeHeaders:    []string{constants.HeaderRequestID},
		AllowCredentials: allowCredentials,
		MaxAge:           12 * time.Hour,
	})
}

func ResponseInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header(constants.HeaderAPIVersion, "v1")
		c.Next()
	}
}

func NotFound(c *gin.Context) {
	RespondError(c, domainerrors.New(domainerrors.KindNotFound, "NOT_FOUND", "route not found"))
}

func MethodNotAllowed(c *gin.Context) {
	RespondError(c, domainerrors.New(domainerrors.KindMethod, "METHOD_NOT_ALLOWED", "method not allowed"))
}
