package jwt

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/phong0206/go_gin_restful/internal/config"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID    int       `json:"user_id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

type JWTService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewJWTService(cfg config.JWTConfig) *JWTService {
	return &JWTService{
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
	}
}

func (s *JWTService) GenerateAccessToken(userID int, email, username string) (string, error) {
	return s.generate(userID, email, username, AccessToken, s.accessSecret, s.accessTTL)
}

func (s *JWTService) GenerateRefreshToken(userID int, email, username string) (string, error) {
	return s.generate(userID, email, username, RefreshToken, s.refreshSecret, s.refreshTTL)
}

func (s *JWTService) ParseAccessToken(token string) (*Claims, error) {
	return s.parse(token, AccessToken, s.accessSecret)
}

func (s *JWTService) ParseRefreshToken(token string) (*Claims, error) {
	return s.parse(token, RefreshToken, s.refreshSecret)
}

func (s *JWTService) generate(userID int, email, username string, tokenType TokenType, secret []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Email:     email,
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   strconv.Itoa(userID),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

func (s *JWTService) parse(token string, expectedType TokenType, secret []byte) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse %s token: %w", expectedType, err)
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid %s token", expectedType)
	}
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("unexpected token type: %s", claims.TokenType)
	}
	return claims, nil
}
