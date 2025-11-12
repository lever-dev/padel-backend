package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lever-dev/padel-backend/internal/entities"
)

type Repository struct {
	connectionURL string
	pool          *pgxpool.Pool
}

func NewRepository(connectionURL string) *Repository {
	return &Repository{connectionURL: connectionURL}
}

func (r *Repository) Connect(ctx context.Context) error {
	p, err := pgxpool.New(ctx, r.connectionURL)
	if err != nil {
		return fmt.Errorf("pgxpool new: %w", err)
	}

	r.pool = p

	return nil
}

func (r *Repository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

func (r *Repository) Create(ctx context.Context, user *entities.User) error {
	if r.pool == nil {
		return fmt.Errorf("not connected to pool")
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}

	d := newDTO(user)

	if d.LastLoginAt != nil {
		t := d.LastLoginAt.UTC()
		d.LastLoginAt = &t
	}

	_, err := r.pool.Exec(
		ctx,
		createUserQuery,
		d.ID,
		d.Nickname,
		d.HashedPassword,
		d.PhoneNumber,
		d.FirstName,
		d.LastName,
		d.CreatedAt,
		nullableTime(d.LastLoginAt),
	)
	if err != nil {
		return fmt.Errorf("exec create user: %w", err)
	}

	return nil
}

const createUserQuery = `
INSERT INTO users(
	id,
	nickname,
	password,
	phone_number,
	first_name,
	last_name,
	created_at,
	last_login_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

func (r *Repository) GetByID(ctx context.Context, userID string) (*entities.User, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("not connected to pool")
	}

	user, err := scan(r.pool.QueryRow(ctx, getUserByIDQuery, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}

	return &user, nil
}

const getUserByIDQuery = `
SELECT
	id,
	nickname,
	password,
	phone_number,
	first_name,
	last_name,
	created_at,
	last_login_at
FROM users
WHERE id = $1
LIMIT 1
`

func (r *Repository) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*entities.User, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("not connected to pool")
	}

	user, err := scan(r.pool.QueryRow(ctx, getUserByPhoneQuery, phoneNumber))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}

	return &user, nil
}

const getUserByPhoneQuery = `
SELECT
	id,
	nickname,
	password,
	phone_number,
	first_name,
	last_name,
	created_at,
	last_login_at
FROM users
WHERE phone_number = $1
LIMIT 1
`

func (r *Repository) GetByNickname(ctx context.Context, nickname string) (entities.User, error) {
	if r.pool == nil {
		return entities.User{}, fmt.Errorf("not connected to pool")
	}

	user, err := scan(r.pool.QueryRow(ctx, getUserByNicknameQuery, nickname))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.User{}, entities.ErrNotFound
		}
		return entities.User{}, fmt.Errorf("scan user: %w", err)
	}

	return user, nil
}

const getUserByNicknameQuery = `
SELECT
	id,
	nickname,
	password,
	phone_number,
	first_name,
	last_name,
	created_at,
	last_login_at
FROM users
WHERE nickname = $1
LIMIT 1
`

func (r *Repository) UpdateLastLogin(ctx context.Context, userID string, lastLogin time.Time) error {
	if r.pool == nil {
		return fmt.Errorf("not connected to pool")
	}

	lastLogin = lastLogin.UTC()

	tag, err := r.pool.Exec(ctx, updateLastLoginQuery, lastLogin, userID)
	if err != nil {
		return fmt.Errorf("exec update last login: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return entities.ErrNotFound
	}

	return nil
}

const updateLastLoginQuery = `
UPDATE users
SET last_login_at = $1
WHERE id = $2
`

type rowScanner interface {
	Scan(dest ...any) error
}

func scan(scanner rowScanner) (entities.User, error) {
	var (
		d         dto
		lastLogin sql.NullTime
	)

	err := scanner.Scan(
		&d.ID,
		&d.Nickname,
		&d.HashedPassword,
		&d.PhoneNumber,
		&d.FirstName,
		&d.LastName,
		&d.CreatedAt,
		&lastLogin,
	)
	if err != nil {
		return entities.User{}, err
	}

	d.CreatedAt = d.CreatedAt.UTC()

	if lastLogin.Valid {
		t := lastLogin.Time.UTC()
		d.LastLoginAt = &t
	}

	return d.toEntity(), nil
}

func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}

	v := t.UTC()
	return v
}
