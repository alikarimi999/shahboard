package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/alikarimi999/shahboard/profileservice/entity"
	"github.com/alikarimi999/shahboard/types"
)

type userRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *userRepo {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) GetByID(ctx context.Context, id types.ObjectId) (*entity.UserInfo, error) {
	query := "SELECT id, email, name, avatar_url, bio, country, created_at, last_active_at FROM users WHERE id = $1"
	row := r.db.QueryRowContext(ctx, query, id)
	var user entity.UserInfo
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.LastActiveAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
func (r *userRepo) Create(ctx context.Context, user *entity.UserInfo) error {
	query := `INSERT INTO users (id, email, name, avatar_url, bio, country, created_at, last_active_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.Name,
		user.AvatarUrl, user.Bio, user.Country, user.CreatedAt, user.LastActiveAt)
	return err
}

func (r *userRepo) Update(ctx context.Context, user *entity.UserInfo) error {
	query := `UPDATE users SET name = $1, avatar_url = $2, bio = $3, country = $4, last_active_at = $5 WHERE id = $6`
	_, err := r.db.ExecContext(ctx, query, user.Name, user.AvatarUrl, user.Bio, user.Country, user.LastActiveAt, user.ID)
	return err
}

func (r *userRepo) UpdateLastActiveAt(ctx context.Context, id types.ObjectId, t time.Time) error {
	query := `UPDATE users SET last_active_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, t, id)
	return err
}
