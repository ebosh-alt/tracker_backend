-- +goose Up
-- Таблица сессий авторизации:
-- хранит хеши access/refresh токенов, их JTI, сроки жизни и статус отзыва.
CREATE TABLE IF NOT EXISTS auth_sessions (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  -- JTI (уникальные идентификаторы) токенов для проверки/ротации.
  access_jti TEXT NOT NULL UNIQUE,
  refresh_jti TEXT NOT NULL UNIQUE,

  -- Храним только хеши токенов (например SHA-256 hex), а не сырые значения.
  access_token_hash TEXT NOT NULL,
  refresh_token_hash TEXT NOT NULL UNIQUE,

  issued_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  access_expires_at TIMESTAMPTZ NOT NULL,
  refresh_expires_at TIMESTAMPTZ NOT NULL,

  -- Ссылка на предыдущую сессию в цепочке ротации refresh.
  rotated_from_id BIGINT REFERENCES auth_sessions(id) ON DELETE SET NULL,

  revoked_at TIMESTAMPTZ,
  revoke_reason TEXT,
  ip TEXT,
  user_agent TEXT,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT auth_sessions_exp_order_chk CHECK (access_expires_at <= refresh_expires_at)
);

CREATE INDEX IF NOT EXISTS auth_sessions_user_idx
  ON auth_sessions(user_id);

CREATE INDEX IF NOT EXISTS auth_sessions_user_active_idx
  ON auth_sessions(user_id, refresh_expires_at)
  WHERE revoked_at IS NULL;

CREATE INDEX IF NOT EXISTS auth_sessions_refresh_active_idx
  ON auth_sessions(refresh_jti, refresh_expires_at)
  WHERE revoked_at IS NULL;

CREATE INDEX IF NOT EXISTS auth_sessions_created_at_idx
  ON auth_sessions(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS auth_sessions;
