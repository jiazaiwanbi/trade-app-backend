package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	authdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/auth"
	userdomain "github.com/jiazaiwanbi/second-hand-platform/internal/domain/user"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user userdomain.User) (userdomain.User, error) {
	query := `
		INSERT INTO users (email, password_hash, nickname, bio)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, password_hash, nickname, bio, created_at, updated_at
	`

	created, err := scanUser(r.pool.QueryRow(ctx, query, user.Email, user.PasswordHash, user.Nickname, user.Bio))
	if err != nil {
		if isUniqueViolation(err) {
			return userdomain.User{}, authdomain.ErrEmailAlreadyExists
		}
		return userdomain.User{}, fmt.Errorf("insert user: %w", err)
	}
	return created, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (userdomain.User, error) {
	query := `
		SELECT id, email, password_hash, nickname, bio, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	user, err := scanUser(r.pool.QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return userdomain.User{}, userdomain.ErrUserNotFound
		}
		return userdomain.User{}, fmt.Errorf("find user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (userdomain.User, error) {
	query := `
		SELECT id, email, password_hash, nickname, bio, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user, err := scanUser(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return userdomain.User{}, userdomain.ErrUserNotFound
		}
		return userdomain.User{}, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id int64, nickname string, bio string) (userdomain.User, error) {
	query := `
		UPDATE users
		SET nickname = $2, bio = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, password_hash, nickname, bio, created_at, updated_at
	`
	user, err := scanUser(r.pool.QueryRow(ctx, query, id, nickname, bio))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return userdomain.User{}, userdomain.ErrUserNotFound
		}
		return userdomain.User{}, fmt.Errorf("update profile: %w", err)
	}
	return user, nil
}

func scanUser(row pgx.Row) (userdomain.User, error) {
	var user userdomain.User
	var createdAt time.Time
	var updatedAt time.Time
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Nickname, &user.Bio, &createdAt, &updatedAt); err != nil {
		return userdomain.User{}, err
	}
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt
	return user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
