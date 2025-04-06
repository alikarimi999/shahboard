package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/alikarimi999/shahboard/profileservice/entity"
	"github.com/alikarimi999/shahboard/profileservice/service/user"
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
	err := row.Scan(&user.ID, &user.Email, &user.Name, &user.AvatarUrl, &user.Bio, &user.Country,
		&user.CreatedAt, &user.LastActiveAt)

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

func (r *userRepo) Update(ctx context.Context, userId types.ObjectId, req user.UpdateUserRequest) error {
	fields := []string{}
	args := []interface{}{}
	argID := 1

	if req.Name != "" {
		fields = append(fields, fmt.Sprintf("name = $%d", argID))
		args = append(args, req.Name)
		argID++
	}
	if req.AvatarUrl != "" {
		fields = append(fields, fmt.Sprintf("avatar_url = $%d", argID))
		args = append(args, req.AvatarUrl)
		argID++
	}
	if req.Bio != "" {
		fields = append(fields, fmt.Sprintf("bio = $%d", argID))
		args = append(args, req.Bio)
		argID++
	}
	if req.Country != "" {
		fields = append(fields, fmt.Sprintf("country = $%d", argID))
		args = append(args, req.Country)
		argID++
	}

	if len(fields) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(fields, ", "), argID)
	args = append(args, userId.String())

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *userRepo) UpdateLastActiveAt(ctx context.Context, id types.ObjectId, t time.Time) error {
	query := `UPDATE users SET last_active_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, t, id)
	return err
}
