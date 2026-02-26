-- +goose Up
CREATE TABLE IF NOT EXISTS workout_sets (
  id BIGSERIAL PRIMARY KEY,
  workout_id BIGINT NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
  order_no INT NOT NULL CHECK (order_no > 0),
  reps INT NOT NULL CHECK (reps > 0),
  weight_kg NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (weight_kg >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS workout_sets_workout_order_idx
  ON workout_sets(workout_id, order_no);

-- +goose Down
DROP INDEX IF EXISTS workout_sets_workout_order_idx;
DROP TABLE IF EXISTS workout_sets;
