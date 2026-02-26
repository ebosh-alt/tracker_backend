-- +goose Up
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS steps_goal INT NOT NULL DEFAULT 10000,
  ADD COLUMN IF NOT EXISTS streak INT NOT NULL DEFAULT 0;

ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_steps_goal_check;

ALTER TABLE users
  ADD CONSTRAINT users_steps_goal_check
  CHECK (steps_goal BETWEEN 1000 AND 30000 AND steps_goal % 500 = 0);

ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_streak_check;

ALTER TABLE users
  ADD CONSTRAINT users_streak_check
  CHECK (streak >= 0);

-- +goose Down
ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_streak_check;

ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_steps_goal_check;

ALTER TABLE users
  DROP COLUMN IF EXISTS streak,
  DROP COLUMN IF EXISTS steps_goal;
