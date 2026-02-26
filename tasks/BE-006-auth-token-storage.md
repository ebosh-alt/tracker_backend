# BE-006 — Auth: хранение токен-сессий в БД

## Метаданные
- ID: `BE-006`
- Приоритет: `P0`
- Статус: `Todo`
- Домен: `identity/auth`

## Цель
Убрать полностью статeless-модель refresh, добавить серверный контроль токен-сессий: хранение хешей токенов, ротация refresh, отзыв сессий и проверка JTI.

## Почему
- Без БД-сессий невозможно надежно отозвать refresh-токен до его `exp`.
- Нельзя детектировать reuse украденного refresh-токена.
- Нельзя реализовать принудительный logout всех устройств.

## Миграция БД
- Добавлена: `/Users/eboshit/projects/tracker/backend/migrations/014_create_auth_sessions.sql`
- Создает таблицу `auth_sessions`:
  - `user_id`,
  - `access_jti`, `refresh_jti`,
  - `access_token_hash`, `refresh_token_hash`,
  - `access_expires_at`, `refresh_expires_at`,
  - `rotated_from_id`, `revoked_at`, `revoke_reason`,
  - `ip`, `user_agent`, `created_at`, `updated_at`.

## План обновления (по слоям)
1. `domain/identity`:
- добавить сущность `Session`,
- инварианты:
  - активная сессия: `revoked_at IS NULL` и `refresh_expires_at > now`,
  - нельзя ротировать уже revoked/expired сессию.
- порты:
  - `SessionWriter`: create/rotate/revoke,
  - `SessionReader`: get-by-refresh-jti/hash, list-active-by-user.

2. `application/auth`:
- обновить `TelegramAuth`:
  - при логине выпускать токены с `jti`,
  - сохранять новую `auth_session` в БД.
- реализовать `RefreshToken` use case:
  - parse refresh JWT,
  - найти сессию по `refresh_jti`,
  - сравнить `refresh_token_hash`,
  - проверить active/expiry,
  - выполнить ротацию: отозвать старую и создать новую сессию,
  - выдать новую пару токенов.
- добавить use case `Logout` и `LogoutAll`.

3. `infra/auth`:
- JWT manager:
  - добавить `jti` в access и refresh claims,
  - вернуть метаданные токенов (`accessJTI`, `refreshJTI`, `exp`).
- hash provider:
  - `SHA-256` (hex/base64), сравнение в constant-time.

4. `infra/postgres`:
- новый репозиторий `auth_session_repo.go`:
  - SQL в `const` в начале файла,
  - `Create`, `GetByRefreshJTI`, `Rotate`, `Revoke`, `RevokeAllByUser`.
- все изменения refresh flow выполнять в `WithinTx`.

5. `delivery/http`:
- `POST /api/auth/telegram`:
  - сохранить сессию после выдачи токенов.
- `POST /api/auth/refresh`:
  - использовать новую ротацию через БД.
- опционально:
  - `POST /api/auth/logout`,
  - `POST /api/auth/logout-all`.

6. `observability/security`:
- логировать `session_id`, `refresh_jti`, причину revoke (без токенов),
- запретить логирование сырых access/refresh токенов.

## Порядок внедрения
1. Применить миграцию `014`.
2. Добавить repo и domain-порты с unit-тестами.
3. Добавить выпуск токенов с `jti` и persist сессии на login.
4. Реализовать refresh-ротацию через БД.
5. Добавить logout/logout-all.
6. Прогнать регрессию auth flow.

## Критерии приемки (DoD)
1. После `POST /api/auth/telegram` создается запись в `auth_sessions`.
2. `POST /api/auth/refresh` работает только для активной и неотозванной сессии.
3. При refresh старая сессия помечается отозванной/rotated.
4. Повторное использование старого refresh-токена возвращает `401`.
5. Реализован отзыв всех сессий пользователя (`logout-all`).
6. Токены в логах и БД хранятся только в виде хеша.
