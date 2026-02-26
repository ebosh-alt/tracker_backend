-- +goose Up
CREATE TABLE IF NOT EXISTS medications (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  dose NUMERIC(10,2) NOT NULL DEFAULT 0,
  unit TEXT NOT NULL DEFAULT 'mg',
  schedule JSONB NOT NULL DEFAULT '{}'::jsonb,
  active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS medications_user_idx ON medications(user_id);
CREATE INDEX IF NOT EXISTS medications_active_idx ON medications(active);

-- +goose Down
DROP INDEX IF EXISTS medications_active_idx;
DROP INDEX IF EXISTS medications_user_idx;
DROP TABLE IF EXISTS medications;
