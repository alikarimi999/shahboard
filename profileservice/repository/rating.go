package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/pkg/paginate"
	pagesql "github.com/alikarimi999/shahboard/pkg/paginate/sql"
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
	err := row.Scan(&rating.UserId, &rating.CurrentScore, &rating.BestScore, &rating.GamesPlayed,
		&rating.GamesWon, &rating.GamesLost, &rating.GamesDraw, &rating.LastUpdated)
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
            INSERT INTO ratings (user_id, current_score, best_score, games_played, games_won, games_lost, games_draw, last_updated)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (user_id)
				DO UPDATE SET 
				current_score = EXCLUDED.current_score,
				best_score = EXCLUDED.best_score,
				games_played = EXCLUDED.games_played,
				games_won = EXCLUDED.games_won,
				games_lost = EXCLUDED.games_lost,
				games_draw = EXCLUDED.games_draw,
				last_updated = EXCLUDED.last_updated
        `
		_, err := tx.ExecContext(ctx, query, r.UserId, r.CurrentScore, r.BestScore, r.GamesPlayed,
			r.GamesWon, r.GamesLost, r.GamesDraw, r.LastUpdated)
		if err != nil {
			return fmt.Errorf("failed to update rating for user %s: %w", r.UserId, err)
		}
	}

	// Insert game Elo changes into the game_elo_changes table
	for _, c := range changes {
		query := `
            INSERT INTO game_elo_changes (user_id, game_id, opponent_id, elo_change, result, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6)
        `
		_, err := tx.ExecContext(ctx, query, c.UserId.String(), c.GameId.String(), c.OpponentId.String(),
			c.EloChange, c.Result, c.UpdatedAt)
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
	query := "SELECT id, user_id, game_id, opponent_id, elo_change, result, updated_at FROM game_elo_changes WHERE user_id = $1"
	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var changes []*entity.GameEloChange
	for rows.Next() {
		var c entity.GameEloChange
		err := rows.Scan(&c.Id, &c.UserId, &c.GameId, &c.OpponentId, &c.EloChange, &c.Result, &c.UpdatedAt)
		if err != nil {
			r.l.Error(fmt.Sprintf("failed to scan row: %v", err))
			continue
		}
		changes = append(changes, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %v", err)
	}

	return changes, nil
}

func (r *ratingRepo) GetGameEloChanges(c context.Context, p *paginate.Paginated) ([]*entity.GameEloChange, uint64, error) {
	limit := p.PerPage
	offset := (p.Page - 1) * limit

	q, cq, args := pagesql.WriteQuery("game_elo_changes", p.Filters, p.SortColumn, p.Decscending, limit, offset)

	rows, err := r.db.QueryContext(c, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	var changes []*entity.GameEloChange
	for rows.Next() {
		var c entity.GameEloChange
		err := rows.Scan(&c.Id, &c.UserId, &c.GameId, &c.OpponentId, &c.EloChange, &c.UpdatedAt, &c.Result)
		if err != nil {
			r.l.Error(fmt.Sprintf("failed to scan row: %v", err))
			continue
		}
		changes = append(changes, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate over rows: %v", err)
	}

	var totalCount int
	cArgs := []interface{}{}
	if len(args) > 0 && len(args) < 3 {
		cArgs = append(cArgs, args[0])
	}
	if len(args) > 0 && len(args) >= 3 {
		cArgs = append(cArgs, args[:len(args)-2]...)
	}

	err = r.db.QueryRowContext(c, cq, cArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute count query: %v", err)
	}

	return changes, uint64(totalCount), nil
}
