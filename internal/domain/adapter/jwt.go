package adapter

import (
	"github.com/phong0206/go_gin_restful/pkg/jwt"
)

type IJWTService interface {
	GenerateAccessToken(userID int, email, username string) (string, error)
	GenerateRefreshToken(userID int, email, username string) (string, error)
	ParseAccessToken(token string) (*jwt.Claims, error)
	ParseRefreshToken(token string) (*jwt.Claims, error)
}