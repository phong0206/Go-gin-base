package usecases

import (
	"context"
	"strings"

	"github.com/phong0206/go_gin_restful/internal/common/helper"
	domainerrors "github.com/phong0206/go_gin_restful/internal/domain/errors"
	"github.com/phong0206/go_gin_restful/internal/domain/repositories"
	"github.com/phong0206/go_gin_restful/internal/dto"
)

type UserUsecase struct {
	users  repositories.IUserRepository
	hasher helper.IPasswordHasher
}

func NewUserUsecase(users repositories.IUserRepository, hasher helper.IPasswordHasher) *UserUsecase {
	return &UserUsecase{users: users, hasher: hasher}
}

func (q *UserUsecase) GetByID(ctx context.Context, id int) (dto.UserResponse, error) {
	if q.users == nil {
		return dto.UserResponse{}, domainerrors.New(domainerrors.KindInternal, "DATABASE_UNAVAILABLE", "database connection is not available")
	}

	user, err := q.users.FindByID(ctx, id)
	if err != nil {
		return dto.UserResponse{}, err
	}
	if user == nil {
		return dto.UserResponse{}, domainerrors.ErrUserNotFound
	}

	return dto.NewUserResponse(user), nil
}

func (q *UserUsecase) Update(ctx context.Context, id int, req dto.UpdateUserRequest) (dto.UserResponse, error) {
	if q.users == nil {
		return dto.UserResponse{}, domainerrors.New(domainerrors.KindInternal, "DATABASE_UNAVAILABLE", "database connection is not available")
	}
	if req.IsActive == nil && req.Password == nil {
		return dto.UserResponse{}, domainerrors.New(domainerrors.KindValidation, "EMPTY_UPDATE", "at least one field must be provided")
	}
	if req.Password != nil && strings.TrimSpace(*req.Password) == "" {
		return dto.UserResponse{}, domainerrors.New(domainerrors.KindValidation, "INVALID_PASSWORD", "password must not be empty")
	}

	var passwordHash *string
	if req.Password != nil {
		if q.hasher == nil {
			return dto.UserResponse{}, domainerrors.New(domainerrors.KindInternal, "PASSWORD_SERVICE_UNAVAILABLE", "password service is not available")
		}
		hash, err := q.hasher.Hash(*req.Password)
		if err != nil {
			return dto.UserResponse{}, domainerrors.Wrap(domainerrors.KindInternal, "PASSWORD_HASH_FAILED", "failed to secure password", err)
		}
		passwordHash = &hash
	}

	user, err := q.users.Update(ctx, id, req.IsActive, passwordHash)
	if err != nil {
		return dto.UserResponse{}, err
	}
	if user == nil {
		return dto.UserResponse{}, domainerrors.ErrUserNotFound
	}

	return dto.NewUserResponse(user), nil
}

func (q *UserUsecase) Delete(ctx context.Context, id int) error {
	if q.users == nil {
		return domainerrors.New(domainerrors.KindInternal, "DATABASE_UNAVAILABLE", "database connection is not available")
	}

	user, err := q.users.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return domainerrors.ErrUserNotFound
	}

	return q.users.SoftDelete(ctx, id)
}
