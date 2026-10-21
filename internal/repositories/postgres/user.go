package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/phong0206/go_gin_restful/internal/domain/entities"
	domainerrors "github.com/phong0206/go_gin_restful/internal/domain/errors"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(ctx context.Context, id int) (*entities.User, error) {
	return r.findOne(ctx, `WHERE id = $1`, id)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	return r.findOne(ctx, `WHERE email = $1`, email)
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	return r.findOne(ctx, `WHERE username = $1`, username)
}

func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	const query = `
		INSERT INTO users (email, username, password_hash, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	if err := r.db.GetContext(ctx, user, query, user.Email, user.Username, user.PasswordHash, user.IsActive); err != nil {
		return mapDatabaseError(err)
	}
	return nil
}

func (r *UserRepository) UpdateRefreshToken(ctx context.Context, userID int, refreshToken string) error {
	const query = `
		UPDATE users
		SET refresh_token = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, refreshToken, userID)
	return err
}

func (r *UserRepository) ClearRefreshToken(ctx context.Context, userID int, refreshToken string) error {
	const query = `
		UPDATE users
		SET refresh_token = NULL, updated_at = NOW()
		WHERE id = $1 AND refresh_token = $2
	`
	_, err := r.db.ExecContext(ctx, query, userID, refreshToken)
	return err
}

func (r *UserRepository) Update(ctx context.Context, userID int, isActive *bool, passwordHash *string) (*entities.User, error) {
	const query = `
		UPDATE users
		SET is_active = COALESCE($2, is_active),
			password_hash = COALESCE($3, password_hash),
			refresh_token = CASE WHEN $2 = FALSE OR $3 IS NOT NULL THEN NULL ELSE refresh_token END,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, username, password_hash, refresh_token, is_active, deleted_at, created_at, updated_at
	`

	var user entities.User
	if err := r.db.GetContext(ctx, &user, query, userID, isActive, passwordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) SoftDelete(ctx context.Context, userID int) error {
	const query = `
		UPDATE users
		SET deleted_at = NOW(), is_active = FALSE, refresh_token = NULL, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func mapDatabaseError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return err
	}

	switch pgErr.ConstraintName {
	case "users_email_key", "users_int_email_key":
		return domainerrors.New(domainerrors.KindConflict, "EMAIL_ALREADY_EXISTS", "email is already in use")
	case "users_username_key", "users_int_username_key":
		return domainerrors.New(domainerrors.KindConflict, "USERNAME_ALREADY_EXISTS", "username is already in use")
	default:
		return domainerrors.New(domainerrors.KindConflict, "USER_ALREADY_EXISTS", "user already exists")
	}
}

func (r *UserRepository) findOne(ctx context.Context, where string, value any) (*entities.User, error) {
	query := `
		SELECT id, email, username, password_hash, refresh_token, is_active, deleted_at, created_at, updated_at
		FROM users
		` + where + ` AND deleted_at IS NULL`

	var user entities.User
	if err := r.db.GetContext(ctx, &user, query, value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
