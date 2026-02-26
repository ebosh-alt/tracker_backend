-- +goose Up
CREATE TABLE IF NOT EXISTS workouts (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exercise_id BIGINT REFERENCES exercises(id) ON DELETE SET NULL,
  started_at TIMESTAMPTZ NOT NULL,
  ended_at TIMESTAMPTZ NOT NULL,
  duration_sec INT NOT NULL CHECK (duration_sec >= 0),
  calories INT NOT NULL DEFAULT 0,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS workouts_user_idx ON workouts(user_id);
CREATE INDEX IF NOT EXISTS workouts_started_idx ON workouts(started_at);

-- +goose Down
DROP INDEX IF EXISTS workouts_started_idx;
DROP INDEX IF EXISTS workouts_user_idx;
DROP TABLE IF EXISTS workouts;
