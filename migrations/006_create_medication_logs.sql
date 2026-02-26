-- +goose Up
CREATE TABLE IF NOT EXISTS medication_logs (
  id BIGSERIAL PRIMARY KEY,
  medication_id BIGINT NOT NULL REFERENCES medications(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  scheduled_at TIMESTAMPTZ NOT NULL,
  taken_at TIMESTAMPTZ,
  status TEXT NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS medication_logs_unique_idx ON medication_logs(medication_id, scheduled_at);
CREATE INDEX IF NOT EXISTS medication_logs_user_idx ON medication_logs(user_id, scheduled_at);

-- +goose Down
DROP INDEX IF EXISTS medication_logs_user_idx;
DROP INDEX IF EXISTS medication_logs_unique_idx;
DROP TABLE IF EXISTS medication_logs;
