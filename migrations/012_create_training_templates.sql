-- +goose Up
CREATE TABLE IF NOT EXISTS training_templates (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS training_template_items (
  id BIGSERIAL PRIMARY KEY,
  template_id BIGINT NOT NULL REFERENCES training_templates(id) ON DELETE CASCADE,
  order_no INT NOT NULL CHECK (order_no > 0),
  exercise_id BIGINT NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
  sets INT NOT NULL DEFAULT 1 CHECK (sets > 0),
  reps INT NOT NULL DEFAULT 1 CHECK (reps > 0),
  weight_kg NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (weight_kg >= 0),
  duration_sec INT NOT NULL DEFAULT 0 CHECK (duration_sec >= 0),
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS training_templates_user_updated_idx
  ON training_templates(user_id, updated_at);

CREATE UNIQUE INDEX IF NOT EXISTS training_template_items_template_order_idx
  ON training_template_items(template_id, order_no);

-- +goose Down
DROP INDEX IF EXISTS training_template_items_template_order_idx;
DROP INDEX IF EXISTS training_templates_user_updated_idx;
DROP TABLE IF EXISTS training_template_items;
DROP TABLE IF EXISTS training_templates;
