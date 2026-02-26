# BE-005 — API: `POST /api/auth/refresh`, `GET /api/me`, `PATCH /api/me/settings`

## Метаданные
- ID: `BE-005`
- Приоритет: `P0`
- Статус: `Todo`
- Домен: `identity/auth`, `user`

## Цель
Закрыть базовый контур авторизации и профиля: после входа через Telegram клиент должен уметь обновлять токены и читать/обновлять пользовательские настройки.

## Контракт
1. `POST /api/auth/refresh`:
- Вход: `{ refreshToken }`
- Выход `200`: `{ token: { accessToken, refreshToken } }`
- Ошибки:
  - `400` — invalid payload
  - `401` — invalid/expired refresh token

2. `GET /api/me`:
- Вход: JWT access token (middleware)
- Выход `200`: `{ id, tgId, username, firstName, lastName, timezone, stepsGoal, streak, createdAt, updatedAt }`
- Ошибки:
  - `401` — unauthorized
  - `404` — user not found

3. `PATCH /api/me/settings`:
- Вход: `{ timezone?, stepsGoal? }` (partial update)
- Выход `200`: обновленный профиль пользователя
- Ошибки:
  - `400` — invalid input (timezone, stepsGoal)
  - `401` — unauthorized
  - `404` — user not found

## Область работ
1. `delivery/http`:
- расширить `authapi` endpoint-ом `POST /auth/refresh` (public route, без JWT middleware),
- добавить `meapi` handler для `GET /me` и `PATCH /me/settings` (protected routes),
- сделать маппинг ATO -> `application.Input` -> `application.Output` -> DTO,
- централизованно отдать ошибки в формате `{code,error}`.

2. `application`:
- use case `RefreshToken`:
  - нормализация входа `refreshToken`,
  - валидация/parse refresh JWT,
  - выпуск новой пары токенов.
- use case `GetMe`:
  - чтение профиля пользователя по `userID` из auth context.
- use case `UpdateMeSettings`:
  - partial update `timezone/stepsGoal`,
  - применение доменной валидации,
  - выполнение через `WithinTx(ctx, func(ctx context.Context) error) error`.

3. `domain/user`:
- использовать инварианты `StepsGoal` и timezone как единую доменную валидацию,
- закрепить контракт изменения настроек через `SettingsPatch`.

4. `infra`:
- `postgres/user_repo`:
  - добавить/довести SQL-методы под read/update профиля,
  - вынести SQL-запросы в `const` в начале файла.
- `infra/auth`:
  - адаптер parse refresh token + issue token pair для application use case.

5. `cmd/api`:
- wiring новых use case и handlers в `main.go`,
- подключение роутов в `server/router.go`:
  - public: `/api/auth/refresh`,
  - protected: `/api/me`, `/api/me/settings`.

6. `tests`:
- unit-тесты use case: `RefreshToken`, `GetMe`, `UpdateMeSettings`,
- handler-тесты для `authapi/meapi` (success + error mapping),
- интеграционные тесты репозитория на чтение и обновление настроек,
- smoke-flow: `auth/telegram -> auth/refresh -> me -> me/settings -> me`.

## Не входит в задачу
- endpoint-ы БАД (`/api/supplements/*`, `/api/medication-logs/ensure-range`)
- endpoint-ы тренировочных шаблонов (`/api/training-templates/*`)
- endpoint-ы дашборда (`/api/dashboard`)

## Критерии приемки (DoD)
1. Реализованы и подключены endpoint-ы:
- `POST /api/auth/refresh`
- `GET /api/me`
- `PATCH /api/me/settings`
2. Все handler-ы остаются тонкими (без SQL, транзакций и доменных правил).
3. Бизнес-валидация выполняется в domain/application слоях.
4. Ошибки отдаются в формате `{code,error}` с корректными HTTP-кодами.
5. Тесты для нового пула API проходят: `go test ./...`.

## Зависимости
- `BE-004` завершен (Telegram auth уже реализован).
- миграции `008_add_users_steps_goal_and_streak.sql` применены.

## Ссылки
- План: `/Users/eboshit/projects/tracker/backend/IMPLEMENTATION_PLAN.md`
- Backlog: `/Users/eboshit/projects/tracker/backend/IMPLEMENTATION_TASKS.md`
- Требования: `/Users/eboshit/projects/tracker/backend/BACKEND_REQUIREMENTS.md`
