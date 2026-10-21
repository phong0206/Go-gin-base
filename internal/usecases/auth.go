package usecases

import (
	"context"
	"strings"

	"github.com/phong0206/go_gin_restful/internal/common/helper"
	"github.com/phong0206/go_gin_restful/internal/domain/adapter"
	"github.com/phong0206/go_gin_restful/internal/domain/entities"
	domainerrors "github.com/phong0206/go_gin_restful/internal/domain/errors"
	"github.com/phong0206/go_gin_restful/internal/domain/repositories"
	"github.com/phong0206/go_gin_restful/internal/dto"
	jwtpkg "github.com/phong0206/go_gin_restful/pkg/jwt"
)

type AuthUsecase struct {
	users  repositories.IUserRepository
	jwt    adapter.IJWTService
	hasher helper.IPasswordHasher
}

func NewAuthUsecase(users repositories.IUserRepository, jwt adapter.IJWTService, hasher helper.IPasswordHasher) *AuthUsecase {
	return &AuthUsecase{
		users:  users,
		jwt:    jwt,
		hasher: hasher,
	}
}

func (u *AuthUsecase) Register(ctx context.Context, req dto.RegisterRequest) (dto.UserResponse, error) {
	if u.users == nil {
		return dto.UserResponse{}, domainerrors.New(domainerrors.KindInternal, "DATABASE_UNAVAILABLE", "database connection is not available")
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := u.users.FindByEmail(ctx, email)
	if err != nil {
		return dto.UserResponse{}, err
	}
	if user != nil {
		return dto.UserResponse{}, domainerrors.New(domainerrors.KindConflict, "EMAIL_ALREADY_EXISTS", "email is already in use")
	}

	passwordHash, err := u.hasher.Hash(req.Password)
	if err != nil {
		return dto.UserResponse{}, domainerrors.Wrap(domainerrors.KindInternal, "PASSWORD_HASH_FAILED", "failed to secure password", err)
	}

	newUser := &entities.User{
		Email:        email,
		PasswordHash: passwordHash,
		IsActive:     true,
		Username:     strings.Split(email, "@")[0],
	}
	if err := u.users.Create(ctx, newUser); err != nil {
		return dto.UserResponse{}, err
	}

	return dto.NewUserResponse(newUser), nil
}

func (u *AuthUsecase) Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
	if u.users == nil {
		return dto.LoginResponse{}, domainerrors.New(domainerrors.KindInternal, "DATABASE_UNAVAILABLE", "database connection is not available")
	}

	user, err := u.users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		return dto.LoginResponse{}, err
	}
	if user == nil {
		return dto.LoginResponse{}, domainerrors.ErrUserNotFound
	}

	if err := u.hasher.Compare(user.PasswordHash, req.Password); err != nil {
		return dto.LoginResponse{}, domainerrors.New(domainerrors.KindUnauthorized, "INVALID_CREDENTIALS", "invalid credentials")
	}

	accessToken, err := u.jwt.GenerateAccessToken(user.ID, user.Email, user.Username)
	if err != nil {
		return dto.LoginResponse{}, err
	}

	refreshToken, err := u.jwt.GenerateRefreshToken(user.ID, user.Email, user.Username)
	if err != nil {
		return dto.LoginResponse{}, err
	}
	if err := u.users.UpdateRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return dto.LoginResponse{}, err
	}

	return dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *AuthUsecase) RefreshToken(ctx context.Context, refreshToken string) (dto.RefreshTokenResponse, error) {
	if u.users == nil {
		return dto.RefreshTokenResponse{}, domainerrors.New(domainerrors.KindInternal, "DATABASE_UNAVAILABLE", "database connection is not available")
	}

	claims, err := u.jwt.ParseRefreshToken(refreshToken)
	if err != nil || claims == nil {
		return dto.RefreshTokenResponse{}, domainerrors.New(domainerrors.KindUnauthorized, "INVALID_REFRESH_TOKEN", "invalid or expired refresh token")
	}

	user, err := findUserByClaims(ctx, u.users, claims)
	if err != nil {
		return dto.RefreshTokenResponse{}, err
	}
	if user == nil || !user.IsActive || user.RefreshToken == nil || *user.RefreshToken != refreshToken {
		return dto.RefreshTokenResponse{}, domainerrors.New(domainerrors.KindUnauthorized, "INVALID_REFRESH_TOKEN", "invalid or revoked refresh token")
	}

	accessToken, err := u.jwt.GenerateAccessToken(user.ID, user.Email, user.Username)
	if err != nil {
		return dto.RefreshTokenResponse{}, err
	}

	return dto.RefreshTokenResponse{AccessToken: accessToken}, nil
}

func (u *AuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	if u.users == nil {
		return domainerrors.New(domainerrors.KindInternal, "DATABASE_UNAVAILABLE", "database connection is not available")
	}

	claims, err := u.jwt.ParseRefreshToken(refreshToken)
	if err != nil || claims == nil {
		return domainerrors.New(domainerrors.KindUnauthorized, "INVALID_REFRESH_TOKEN", "invalid or expired refresh token")
	}

	user, err := findUserByClaims(ctx, u.users, claims)
	if err != nil {
		return err
	}
	if user == nil || !user.IsActive || user.RefreshToken == nil || *user.RefreshToken != refreshToken {
		return domainerrors.New(domainerrors.KindUnauthorized, "INVALID_REFRESH_TOKEN", "invalid or revoked refresh token")
	}

	return u.users.ClearRefreshToken(ctx, user.ID, refreshToken)
}

func findUserByClaims(ctx context.Context, users repositories.IUserRepository, claims *jwtpkg.Claims) (*entities.User, error) {
	if claims.Email != "" {
		user, err := users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(claims.Email)))
		if err != nil {
			return nil, err
		}
		if user != nil {
			return user, nil
		}
	}
	if claims.Username != "" {
		return users.FindByUsername(ctx, claims.Username)
	}
	return nil, nil
}
