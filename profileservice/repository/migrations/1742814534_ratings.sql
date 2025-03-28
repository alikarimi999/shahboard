CREATE TABLE ratings (
    user_id VARCHAR(64) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    current_score BIGINT NOT NULL DEFAULT 1000,
    best_score BIGINT NOT NULL DEFAULT 1000,
    games_played BIGINT NOT NULL DEFAULT 0,
    games_won BIGINT NOT NULL DEFAULT 0,
    games_lost BIGINT NOT NULL DEFAULT 0,
    games_draw BIGINT NOT NULL DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);


CREATE TABLE game_elo_changes (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(64),
    game_id VARCHAR(64),
    opponent_id VARCHAR(64),
    elo_change BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() 
);

CREATE INDEX idx_game_elo_changes_user_id ON game_elo_changes(user_id);