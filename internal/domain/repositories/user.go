package repositories

import (
	"context"

	"github.com/phong0206/go_gin_restful/internal/domain/entities"
)

type IUserRepository interface {
	FindByID(ctx context.Context, id int) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByUsername(ctx context.Context, username string) (*entities.User, error)
	Create(ctx context.Context, user *entities.User) error
	UpdateRefreshToken(ctx context.Context, userID int, refreshToken string) error
	ClearRefreshToken(ctx context.Context, userID int, refreshToken string) error
}
