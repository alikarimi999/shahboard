ALTER TABLE game_elo_changes 
ADD COLUMN result SMALLINT CHECK (result IN (-1, 0, 1));
