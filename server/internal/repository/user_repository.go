package repository

import (
	"context"
	"database/sql"

	"github.com/lnardon/arete/internal/database"
	"github.com/lnardon/arete/internal/models"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, username, passwordHash string) (models.User, error) {
	var u models.User
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, username, password_hash, created_at`,
		username, passwordHash,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)

	return u, err
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	var u models.User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, created_at FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)

	if err == sql.ErrNoRows {
		return u, ErrNotFound
	}
	return u, err
}
