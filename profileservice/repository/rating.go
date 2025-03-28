package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/profileservice/entity"
	"github.com/alikarimi999/shahboard/types"
)

type ratingRepo struct {
	db *sql.DB
	l  log.Logger
}

func NewRatingRepo(db *sql.DB, l log.Logger) *ratingRepo {
	return &ratingRepo{
		db: db,
		l:  l,
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

func (r *ratingRepo) Update(ctx context.Context, ratings []*entity.Rating, changes []*entity.GameEloChange) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Update ratings table
	for _, r := range ratings {
		query := `
            UPDATE ratings
            SET 
                current_score = $1,
                best_score = $2,
                games_played = games_played + 1,
                games_won = games_won + $3,
                games_lost = games_lost + $4,
                games_draw = games_draw + $5,
                last_updated = $6
            WHERE user_id = $7
        `
		_, err := tx.ExecContext(ctx, query, r.CurrentScore, r.BestScore,
			r.GamesWon, r.GamesLost, r.GamesDraw, r.LastUpdated, r.UserId.String())
		if err != nil {
			return fmt.Errorf("failed to update rating for user %s: %w", r.UserId, err)
		}
	}

	// Insert game Elo changes into the game_elo_changes table
	for _, c := range changes {
		query := `
            INSERT INTO game_elo_changes (user_id, game_id, opponent_id, elo_change, updated_at)
            VALUES ($1, $2, $3, $4, $5)
        `
		_, err := tx.ExecContext(ctx, query, c.UserId.String(), c.GameId.String(), c.OpponentId.String(),
			c.EloChange, c.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert Elo change for game %s: %w", c.GameId, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *ratingRepo) GetGameEloChangesByUserId(ctx context.Context, userId types.ObjectId) ([]*entity.GameEloChange, error) {
	query := "SELECT id, user_id, game_id, opponent_id, elo_change, updated_at FROM game_elo_changes WHERE user_id = $1"
	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var changes []*entity.GameEloChange
	for rows.Next() {
		var c entity.GameEloChange
		err := rows.Scan(&c.Id, &c.UserId, &c.GameId, &c.OpponentId, &c.EloChange, &c.UpdatedAt)
		if err != nil {
			r.l.Error(fmt.Sprintf("failed to scan row: %v", err))
			continue
		}
		changes = append(changes, &c)
	}

	return changes, nil
}
