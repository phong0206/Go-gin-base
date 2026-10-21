package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/phong0206/go_gin_restful/internal/common/constants"
	"github.com/phong0206/go_gin_restful/internal/domain/adapter"
	"github.com/phong0206/go_gin_restful/internal/domain/entities"
	domainerrors "github.com/phong0206/go_gin_restful/internal/domain/errors"
	"github.com/phong0206/go_gin_restful/internal/domain/repositories"
)

func VerifyToken(jwtService adapter.IJWTService, users repositories.IUserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if jwtService == nil {
			RespondError(c, domainerrors.New(domainerrors.KindInternal, "JWT_UNAVAILABLE", "token service is not available"))
			return
		}

		token := ExtractToken(c, constants.CookieAccessToken)
		if token == "" {
			RespondError(c, domainerrors.New(domainerrors.KindUnauthorized, "MISSING_ACCESS_TOKEN", "access token is required"))
			return
		}

		claims, err := jwtService.ParseAccessToken(token)
		if err != nil || claims == nil {
			RespondError(c, domainerrors.New(domainerrors.KindUnauthorized, "INVALID_ACCESS_TOKEN", "invalid or expired access token"))
			return
		}

		user, err := findUser(c, users, claims.Email, claims.Username)
		if err != nil {
			RespondError(c, err)
			return
		}
		if user == nil || !user.IsActive {
			RespondError(c, domainerrors.New(domainerrors.KindUnauthorized, "UNAUTHORIZED", "user is not authorized"))
			return
		}

		c.Set(constants.ContextUserKey, user)
		c.Next()
	}
}

func ExtractToken(c *gin.Context, cookieName string) string {
	authorization := strings.Fields(strings.TrimSpace(c.GetHeader("Authorization")))
	if len(authorization) == 2 && strings.EqualFold(authorization[0], "Bearer") {
		return authorization[1]
	}

	token, _ := c.Cookie(cookieName)
	return strings.TrimSpace(token)
}

func ExtractRefreshToken(c *gin.Context) string {
	if token, err := c.Cookie(constants.CookieRefreshToken); err == nil && strings.TrimSpace(token) != "" {
		return strings.TrimSpace(token)
	}
	return ExtractToken(c, constants.CookieRefreshToken)
}

func findUser(c *gin.Context, users repositories.IUserRepository, email, username string) (*entities.User, error) {
	if users == nil {
		return nil, domainerrors.New(domainerrors.KindInternal, "DATABASE_UNAVAILABLE", "database connection is not available")
	}

	if email != "" {
		user, err := users.FindByEmail(c.Request.Context(), strings.ToLower(strings.TrimSpace(email)))
		if err != nil {
			return nil, err
		}
		if user != nil {
			return user, nil
		}
	}
	if username != "" {
		return users.FindByUsername(c.Request.Context(), username)
	}

	return nil, nil
}
