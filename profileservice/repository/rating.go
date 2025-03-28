package repository

import (
	"context"
	"database/sql"

	"github.com/alikarimi999/shahboard/profileservice/entity"
	"github.com/alikarimi999/shahboard/types"
)

type ratingRepo struct {
	db *sql.DB
}

func NewRatingRepo(db *sql.DB) *ratingRepo {
	return &ratingRepo{
		db: db,
	}
}

func (r *ratingRepo) GetByUserId(ctx context.Context, id types.ObjectId) (*entity.Rating, error) {
	query := "SELECT user_id, current_score, best_score, games_played, games_won, games_lost, games_draw, last_updated FROM ratings WHERE user_id = $1"
	row := r.db.QueryRowContext(ctx, query, id)
	var rating entity.Rating
	err := row.Scan(&rating.UserId, &rating.CurrentScore, &rating.BestScore, &rating.GamesPlayed, &rating.GamesWon, &rating.GamesLost, &rating.GamesDraw, &rating.LastUpdated)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rating, nil
}

func (r *ratingRepo) Update(ctx context.Context, ratings ...*entity.Rating) error {
	query := `UPDATE ratings SET current_score = $1, best_score = $2, games_played = $3, games_won = $4, games_lost = $5, games_draw = $6, last_updated = $7 WHERE user_id = $8`
	for _, rating := range ratings {
		_, err := r.db.ExecContext(ctx, query, rating.CurrentScore, rating.BestScore, rating.GamesPlayed, rating.GamesWon, rating.GamesLost, rating.GamesDraw, rating.LastUpdated, rating.UserId)
		if err != nil {
			return err
		}
	}
	return nil
}
