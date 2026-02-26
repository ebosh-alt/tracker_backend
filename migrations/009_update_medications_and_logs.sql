-- +goose Up
ALTER TABLE medications
  ADD COLUMN IF NOT EXISTS color TEXT NOT NULL DEFAULT 'info',
  ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT 'tablet';

ALTER TABLE medications
  DROP CONSTRAINT IF EXISTS medications_color_check;

ALTER TABLE medications
  ADD CONSTRAINT medications_color_check
  CHECK (color IN ('lime', 'violet', 'rose', 'info'));

ALTER TABLE medications
  DROP CONSTRAINT IF EXISTS medications_category_check;

ALTER TABLE medications
  ADD CONSTRAINT medications_category_check
  CHECK (category IN ('tablet', 'sport_nutrition'));

UPDATE medication_logs
SET status = 'missed'
WHERE status = 'skipped';

ALTER TABLE medication_logs
  DROP CONSTRAINT IF EXISTS medication_logs_status_check;

ALTER TABLE medication_logs
  ADD CONSTRAINT medication_logs_status_check
  CHECK (status IN ('pending', 'taken', 'missed'));

-- +goose Down
ALTER TABLE medication_logs
  DROP CONSTRAINT IF EXISTS medication_logs_status_check;

UPDATE medication_logs
SET status = 'skipped'
WHERE status = 'missed';

ALTER TABLE medications
  DROP CONSTRAINT IF EXISTS medications_category_check;

ALTER TABLE medications
  DROP CONSTRAINT IF EXISTS medications_color_check;

ALTER TABLE medications
  DROP COLUMN IF EXISTS category,
  DROP COLUMN IF EXISTS color;
