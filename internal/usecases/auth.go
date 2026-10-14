package usecases

import (
	"context"

	"github.com/phong0206/go_gin_restful/internal/common/helper"
	"github.com/phong0206/go_gin_restful/internal/domain/adapter"
	domainerrors "github.com/phong0206/go_gin_restful/internal/domain/errors"
	"github.com/phong0206/go_gin_restful/internal/domain/repositories"
	"github.com/phong0206/go_gin_restful/internal/dto"
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

func (u *AuthUsecase) Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
	if u.users == nil {
		return dto.LoginResponse{}, domainerrors.New(domainerrors.KindInternal, "DATABASE_UNAVAILABLE", "database connection is not available")
	}

	user, err := u.users.FindByEmail(ctx, req.Email)
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

	return dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
