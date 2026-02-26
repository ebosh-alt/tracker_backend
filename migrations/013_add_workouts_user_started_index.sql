-- +goose Up
CREATE INDEX IF NOT EXISTS workouts_user_started_idx
  ON workouts(user_id, started_at);

-- +goose Down
DROP INDEX IF EXISTS workouts_user_started_idx;
