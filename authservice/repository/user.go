package repository

import (
	"context"
	"database/sql"

	"github.com/alikarimi999/shahboard/authservice/entity"
	"github.com/alikarimi999/shahboard/authservice/service"
)

type userRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) service.Repository {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) Create(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (id, email, created_at, updated_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, user.ID.String(), user.Email, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := "SELECT id, email, created_at, updated_at FROM users WHERE email = $1"
	row := r.db.QueryRowContext(ctx, query, email)
	var user entity.User
	err := row.Scan(&user.ID, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil

}
