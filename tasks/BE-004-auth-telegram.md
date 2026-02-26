# BE-004 — API: `POST /api/auth/telegram`

## Метаданные
- ID: `BE-004`
- Приоритет: `P0`
- Статус: `Done`
- Домен: `identity/auth`

## Цель
Реализовать вход через Telegram как первый блокирующий API-метод для остального защищенного пула.

## Контракт
- Метод: `POST /api/auth/telegram`
- Вход: `{ initData }`
- Опциональный header: `X-Timezone`
- Выход `200`: `{ user, token: { accessToken, refreshToken } }`
- Ошибки:
  - `400` — invalid payload
  - `401` — invalid telegram signature

## Область работ
1. `delivery/http`:
- добавить `authapi` handler и подключение роута без JWT middleware.
2. `application`:
- use case `TelegramAuth`:
  - нормализация входа (`initData`),
  - верификация подписи Telegram,
  - upsert пользователя,
  - выпуск пары токенов.
3. `domain/identity`:
- порты для verify Telegram payload и upsert/read пользователя.
4. `infra`:
- адаптер verify Telegram,
- репозиторий пользователя (postgres),
- выпуск JWT access/refresh.

## Не входит в задачу
- refresh endpoint (`POST /api/auth/refresh`)
- профиль `GET /api/me`
- расширения OpenAPI/Swagger (по текущему приоритету команды)

## Критерии приемки (DoD)
1. Метод принимает: `initData`.
2. На успех возвращает `user` и `token(accessToken, refreshToken)`.
3. Ошибки отдаются в формате `{code,error}`.
4. Добавлены unit-тесты use case и integration-тесты handler.
5. В логах отсутствуют токены и чувствительные данные.

## Зависимости
- Конфигурация JWT секретов в `internal/infra/config/config.dev.yml` / `config.prod.yml`.
- Доступный Telegram verify adapter в infra-слое.

## Ссылки
- План: `/Users/eboshit/projects/tracker/backend/IMPLEMENTATION_PLAN.md`
- Backlog: `/Users/eboshit/projects/tracker/backend/IMPLEMENTATION_TASKS.md`
- Требования: `/Users/eboshit/projects/tracker/backend/BACKEND_REQUIREMENTS.md`
