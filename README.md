# Backend

Go API for Telegram Mini App.

## Local run
1. Start Postgres: `docker compose -f /Users/eboshit/projects/tracker/infra/docker-compose.yml up -d`
2. Fill secrets in `/Users/eboshit/projects/tracker/backend/internal/infra/config/config.dev.yml` (`Telegram.*`, `Auth.*`).
3. Run migrations: `cd /Users/eboshit/projects/tracker/backend && go run ./cmd/migrate up`
4. Run API: `cd /Users/eboshit/projects/tracker/backend && go run ./cmd/api`

## Config
- Config format: `.yml`.
- Default file (when running from `backend`): `/Users/eboshit/projects/tracker/backend/internal/infra/config/config.dev.yml`
- Production file: `/Users/eboshit/projects/tracker/backend/internal/infra/config/config.prod.yml`
- Select environment file via env: `CONFIG_ENV=dev|prod` (default `dev`)
- Override path via env: `CONFIG_PATH=/abs/path/to/config.yml`
- Loader: Viper with strict unmarshal (`UnmarshalExact`).
- Log level: `App.logLevel` (`debug|info|warn|error`)

## Swagger UI
- `GET /docs` (Swagger UI)
- `GET /docs/openapi.json` (spec)

## Auth
All protected endpoints require:

`Authorization: Bearer <access_token>`

`POST /api/auth/telegram` expects `{ "initData": "..." }`.

Optional header: `X-Timezone: Europe/Moscow` (defaults to `UTC`).

Response includes `user` and `token` with `accessToken` + `refreshToken`.

`POST /api/auth/refresh` expects `{ "refreshToken": "..." }`.

## Medication schedule JSON
We store schedule in `medications.schedule` (JSONB). Example:

```json
{
  "byDay": ["MO", "WE", "FR"],
  "times": ["08:00", "21:30"]
}
```

Times are interpreted in the user's timezone.

## Worker
Run notification worker (every 5 minutes):

`cd /Users/eboshit/projects/tracker/backend && go run ./cmd/worker`

## Telegram webhook
`POST /api/telegram/webhook` expects updates from Telegram.

Header required: `X-Telegram-Bot-Api-Secret-Token: <TELEGRAM_WEBHOOK_SECRET>`

Set webhook:
`POST /api/telegram/set-webhook` with `{ "url": "https://your-domain/api/telegram/webhook" }`.

## Medication logs
- `GET /api/medication-logs?from=RFC3339&to=RFC3339&medicationId=ID&status=taken|skipped|pending&limit=50&offset=0`
- `POST /api/medication-logs` with `{ "medicationId": 1, "scheduledAt": "2026-02-22T12:00:00Z", "status": "taken|skipped|pending" }`
- `PUT /api/medication-logs/:id` with `{ "status": "taken|skipped|pending" }`

## Exercises
- Default catalog (15 RU exercises) is auto-seeded on Telegram auth/login.
- `GET /api/exercises?limit=50&offset=0&q=search&sortBy=created_at|name&sortDir=asc|desc`
- `POST /api/exercises`
- `PUT /api/exercises/:id`
- `DELETE /api/exercises/:id`

## Workouts
- `GET /api/workouts?from=RFC3339&to=RFC3339&exerciseId=ID&limit=50&offset=0&sortBy=started_at|created_at&sortDir=asc|desc`
- `POST /api/workouts` body supports `sets`, `reps`, `weightKg`, `calories`, `notes`
- `PUT /api/workouts/:id` body supports `sets`, `reps`, `weightKg`, `calories`, `notes`
- `DELETE /api/workouts/:id`

## Stats
`GET /api/stats?range=day|week|month`
`GET /api/stats?from=YYYY-MM-DD&to=YYYY-MM-DD`

Range is computed in the user's timezone. Output includes totals for steps, workouts, and medications.

Stats includes daily series and analytics by exercise type:

```json
{
  "exerciseTypes": [
    { "type": "cardio", "count": 3, "durationSec": 3600, "calories": 600 }
  ],
  "series": [
    {
      "date": "2026-02-22",
      "steps": 5000,
      "workoutCount": 1,
      "workoutDuration": 1200,
      "workoutCalories": 150,
      "medTaken": 1,
      "medSkipped": 0,
      "medPending": 0
    }
  ]
}
```
# tracker_backend
