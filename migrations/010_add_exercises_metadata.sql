-- +goose Up
ALTER TABLE exercises
  ADD COLUMN IF NOT EXISTS muscle_group TEXT,
  ADD COLUMN IF NOT EXISTS emoji TEXT;

CREATE INDEX IF NOT EXISTS exercises_user_lower_name_idx
  ON exercises (user_id, lower(name));

-- +goose Down
DROP INDEX IF EXISTS exercises_user_lower_name_idx;

ALTER TABLE exercises
  DROP COLUMN IF EXISTS emoji,
  DROP COLUMN IF EXISTS muscle_group;
