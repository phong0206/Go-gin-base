package usecases

import (
	"context"

	domainerrors "github.com/phong0206/go_gin_restful/internal/domain/errors"
	"github.com/phong0206/go_gin_restful/internal/domain/repositories"
	"github.com/phong0206/go_gin_restful/internal/dto"
)

type UserUsecase struct {
	users repositories.IUserRepository
}

func NewUserUsecase(users repositories.IUserRepository) *UserUsecase {
	return &UserUsecase{users: users}
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
