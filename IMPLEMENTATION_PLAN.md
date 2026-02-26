# План реализации Backend

## Контекст

Источник требований:
- `/Users/eboshit/projects/tracker/backend/BACKEND_REQUIREMENTS.md`

Базовая дата плана:
- 24 февраля 2026

## Фаза 0: Синхронизация и дизайн (24-26 февраля 2026)

1. Зафиксировать продуктовую семантику:
- маппинг статусов: `pending | taken | missed` (принимать `skipped` как backward-compatible alias),
- категории БАД: `tablet | sport_nutrition`,
- разделение модели тренировок: `шаблон тренировки` vs `история выполнения`.
- auth-семантика: вход через Telegram `initData` и выдача access/refresh токенов в ответе.

2. Зафиксировать API-контракты:
- финализировать request/response для новых endpoint:
  - `/api/auth/telegram` (payload `initData`),
  - `/api/me/settings`,
  - `/api/steps/analytics`,
  - `/api/supplements/today`,
  - `/api/supplements/calendar`,
  - `/api/supplements/day`,
  - `/api/training-templates` (CRUD),
  - `/api/workouts/activity`,
  - `/api/dashboard`.

3. Артефакты:
- обновленный `/Users/eboshit/projects/tracker/openapi.yaml`,
- документ с планом миграций БД.

4. Критерии выхода:
- нет открытых неоднозначностей в контрактах,
- фронтенд и backend синхронизированы по payload-полям и enum.

## Фаза 1: Core Parity (Спринт 1, 27 февраля-5 марта 2026)

1. Модель данных и миграции:
- добавить `users.steps_goal`,
- добавить `medications.category`,
- убедиться, что `medications.color` присутствует,
- добавить `auth_sessions` для серверного контроля токен-сессий
  (refresh rotation, revoke, session tracking).

2. Реализация API:
- `PATCH /api/me/settings` (`timezone`, `stepsGoal`),
- обновить `GET /api/me` (добавить `stepsGoal` и `streak`),
- добавить `POST /api/auth/refresh` на базе БД-сессий,
- добавить `POST /api/steps/add` (атомарный delta),
- добавить `GET /api/steps/analytics?month=YYYY-MM`.

3. Доменная логика:
- вычисление выполнения плана за неделю и месяц:
  - `weekGoal = stepsGoal * 7`,
  - `monthGoal = stepsGoal * daysInMonth`,
- поддержка quick add для 500/1000/2000/5000/10000,
- ротация refresh-токена:
  - предыдущая сессия помечается `revoked/rotated`,
  - новый refresh выпускается только через новую активную сессию.

4. Тесты:
- unit-тесты на вычисления шаговой аналитики,
- integration-тесты для flow настроек + добавления шагов,
- integration-тесты auth flow: `telegram auth -> refresh -> reject old refresh`.

5. Критерии выхода:
- все endpoint Фазы 1 задокументированы и покрыты тестами,
- обратная совместимость текущих endpoint сохранена.
- endpoint авторизации корректно работает с `initData`,
- refresh reuse блокируется (401), logout/logout-all поддерживают отзыв сессий.

## Фаза 2: Полнота UX по БАД (Спринт 2, 6-12 марта 2026)

1. Реализация API:
- `GET /api/supplements/today?date=YYYY-MM-DD`,
- `GET /api/supplements/calendar?month=YYYY-MM`,
- `GET /api/supplements/day?date=YYYY-MM-DD`,
- `POST /api/medication-logs/ensure-range`.

2. Доменная логика:
- генерация pending-логов по расписанию в запрошенных диапазонах,
- разделение по категориям во всех supplement-ответах,
- day drill-down поведение:
  - для прошедших дней: taken/missed по категориям,
  - для будущих дней: planned записи по категориям.

3. Совместимость:
- принимать `skipped` на входе, отдавать канонический `missed`.

4. Тесты:
- integration-тесты month calendar и day drill-down,
- тест идемпотентности ensure-range.

5. Критерии выхода:
- вкладки БАД полностью питаются backend-контрактами без client-side синтеза данных.

## Фаза 3: Полнота по тренировкам и шаблонам (Спринт 3, 13-19 марта 2026)

1. Модель данных и миграции:
- добавить таблицу `workout_sets`,
- добавить `training_templates`,
- добавить `training_template_items`.

2. Реализация API:
- расширить create/update/get workouts поддержкой массива подходов,
- добавить CRUD шаблонов тренировок:
  - `GET /api/training-templates`,
  - `POST /api/training-templates`,
  - `PUT /api/training-templates/{id}`,
  - `DELETE /api/training-templates/{id}`,
- добавить `GET /api/workouts/activity?month=YYYY-MM`.

3. Каталог упражнений:
- поддержать динамические категории на основе данных упражнений,
- включить UI-метаданные (`muscleGroup`, `emoji`) в контракт.

4. Тесты:
- транзакционные тесты для записи workout + workout_sets,
- тесты CRUD шаблонов + валидации.

5. Критерии выхода:
- все сценарии тренировок и истории доступны через backend-контракты.

## Фаза 4: Дашборд и hardening (Спринт 4, 20-26 марта 2026)

1. Реализация API:
- `GET /api/dashboard?date=YYYY-MM-DD` (агрегирующий endpoint).

2. Производительность и надежность:
- оптимизировать тяжелые stats/calendar-запросы,
- добавить недостающие индексы из требований,
- централизованно применить ограничения пагинации и валидации входа.

3. Наблюдаемость:
- метрики latency и error rate по endpoint,
- структурные логи с request id, user id, route, error code.

4. Тесты:
- contract-тесты по OpenAPI,
- регрессионный набор для критических flow.

5. Критерии выхода:
- p95-таргеты достигнуты на staging,
- нет P1/P2-регрессий.

## Фаза 5: Релиз и стабилизация (Спринт 5, 27 марта-2 апреля 2026)

1. Релизный процесс:
- выкатить миграции БД,
- выкатить backend (при необходимости с feature flags),
- canary rollout и мониторинг error budget.

2. Проверка backward compatibility:
- существующий фронтенд и legacy-клиенты продолжают работать.

3. Финальные артефакты:
- финализированный OpenAPI,
- обновленный runbook и эксплуатационный checklist.

4. Критерии выхода:
- релиз в production завершен,
- мониторинг стабилен в течение 72 часов.

## Декомпозиция работ (WBS)

1. Контракты:
- обновить `/Users/eboshit/projects/tracker/openapi.yaml`,
- регенерировать `/Users/eboshit/projects/tracker/openapi.json`.

2. БД:
- создать новые миграции в `/Users/eboshit/projects/tracker/backend/migrations`.

3. Delivery-слой:
- обновить handlers/ATO/DTO/mappers в:
  - `/Users/eboshit/projects/tracker/backend/internal/delivery/http/server/handlers`,
  - `/Users/eboshit/projects/tracker/backend/internal/delivery/http/server/ato`,
  - `/Users/eboshit/projects/tracker/backend/internal/delivery/http/server/dto`,
  - `/Users/eboshit/projects/tracker/backend/internal/delivery/http/server/mappers`.

4. Usecase и domain:
- обновить:
  - `/Users/eboshit/projects/tracker/backend/internal/usecase`,
  - `/Users/eboshit/projects/tracker/backend/internal/domain`,
  - `/Users/eboshit/projects/tracker/backend/internal/entities`.

5. Repository:
- реализовать SQL-изменения в:
  - `/Users/eboshit/projects/tracker/backend/internal/repository/postgres`.

6. Тесты:
- unit + integration в:
  - `/Users/eboshit/projects/tracker/backend/internal/usecase`,
  - `/Users/eboshit/projects/tracker/backend/internal/tests`.

## Риски и меры снижения

1. Риск:
- ошибки таймзон в календаре и генерации расписаний.
Митигатор:
- фиксированная матрица таймзон в тестах (UTC, Europe/Moscow, Asia/Almaty, America/New_York).

2. Риск:
- рассинхрон контрактов фронтенда и backend в процессе поэтапного релиза.
Митигатор:
- OpenAPI как единый источник истины + contract-тесты в CI.

3. Риск:
- деградация производительности на monthly calendar и dashboard endpoint.
Митигатор:
- pre-aggregation или оптимизированные индексы; нагрузочное тестирование до релиза.

## Definition of Done

1. Все endpoint из требований реализованы и задокументированы.
2. Миграции БД применяются и откатываются.
3. Unit/integration/contract тесты проходят в CI.
4. p95 latency-таргеты достигаются на staging.
5. Frontend flow работают end-to-end без mock/fallback-данных.
