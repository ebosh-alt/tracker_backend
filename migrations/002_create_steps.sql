-- +goose Up
CREATE TABLE IF NOT EXISTS steps (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  date DATE NOT NULL,
  count INT NOT NULL CHECK (count >= 0),
  source TEXT NOT NULL DEFAULT 'manual',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS steps_user_date_idx ON steps(user_id, date);

-- +goose Down
DROP INDEX IF EXISTS steps_user_date_idx;
DROP TABLE IF EXISTS steps;
