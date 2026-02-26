# Backlog задач Backend (готов к исполнению)

Источник:
- `/Users/eboshit/projects/tracker/backend/BACKEND_REQUIREMENTS.md`
- `/Users/eboshit/projects/tracker/backend/IMPLEMENTATION_PLAN.md`

Единица оценки:
- инженерные дни (`d`), целевой размер задачи: `1-2d`.

## Milestone M0: Контракты и дизайн

## BE-001: Финализировать структуру OpenAPI v2
- Estimate: 1d
- Depends on: none
- Scope:
  - описать все новые endpoint и схемы в `openapi.yaml`,
  - синхронизировать enum: `missed`, `tablet|sport_nutrition`.
- Acceptance:
  - OpenAPI проходит валидацию,
  - в paths/schemas нет незакрытых TODO.

## BE-002: Унифицировать контракт ошибок
- Estimate: 1d
- Depends on: BE-001
- Scope:
  - проверить единый маппинг ошибок во всех handlers в `{code,error}`,
  - задокументировать матрицу ошибок по endpoint.
- Acceptance:
  - integration-проверки подтверждают консистентные status/code.

## BE-003: Дизайн-док миграций БД
- Estimate: 1d
- Depends on: BE-001
- Scope:
  - детализировать schema changes для `users.steps_goal`, `medications.category`,
    `workout_sets`, `training_templates`, `training_template_items`.
- Acceptance:
  - план миграций с rollback-стратегией согласован.

## BE-004: Контракт авторизации Telegram `initData`
- Estimate: 1d
- Depends on: BE-001
- Scope:
  - зафиксировать в OpenAPI входной payload через `initData`,
  - явно задокументировать, что в ответе сразу возвращаются `accessToken/refreshToken`.
- Acceptance:
  - контракт `/api/auth/telegram` однозначен и покрыт примерами request/response.

## Milestone M1: Core Parity (настройки + аналитика шагов)

## BE-010: Миграция - users.steps_goal
- Estimate: 1d
- Depends on: BE-003
- Scope:
  - добавить `steps_goal` с default `10000` и check-ограничениями.
- Acceptance:
  - миграция up/down выполняется локально.

## BE-011: Миграция - medications.category
- Estimate: 1d
- Depends on: BE-003
- Scope:
  - добавить `category` с default `tablet`,
  - добавить constraint на допустимые значения.
- Acceptance:
  - недопустимое значение категории отвергается на уровне БД.

## BE-012: Расширить user entity/repo для настроек
- Estimate: 1d
- Depends on: BE-010
- Scope:
  - включить `stepsGoal` в user read model,
  - добавить repo-метод для обновления `timezone` и `steps_goal`.
- Acceptance:
  - repo-тесты подтверждают update и readback.

## BE-013: API - PATCH /api/me/settings
- Estimate: 1d
- Depends on: BE-012
- Scope:
  - добавить handler + mapper + usecase flow,
  - поддержать partial update (`timezone`, `stepsGoal`).
- Acceptance:
  - endpoint обновляет поля атомарно.

## BE-014: API - обогатить GET /api/me полями stepsGoal/streak
- Estimate: 1d
- Depends on: BE-012
- Scope:
  - обновить DTO и mapper,
  - сохранить обратную совместимость существующих полей.
- Acceptance:
  - contract-тест валидирует новый response payload.

## BE-015: API - POST /api/steps/add (атомарный delta)
- Estimate: 1d
- Depends on: BE-010
- Scope:
  - реализовать endpoint инкремента,
  - валидировать ненулевой delta.
- Acceptance:
  - конкурентные запросы дают консистентный итог.

## BE-016: Домен - калькулятор аналитики шагов (week/month)
- Estimate: 1d
- Depends on: BE-010
- Scope:
  - вычислять week/month totals и completion percentages,
  - учитывать границы дней в timezone пользователя.
- Acceptance:
  - детерминированные unit-тесты для пограничных дат и переходов месяца.

## BE-017: API - GET /api/steps/analytics
- Estimate: 1d
- Depends on: BE-016
- Scope:
  - отдавать 7d и month series + выполнение плана,
  - поддержать query `month=YYYY-MM`.
- Acceptance:
  - integration-тест возвращает ожидаемую схему и значения.

## BE-018: Регрессионные тесты core parity
- Estimate: 1d
- Depends on: BE-013, BE-014, BE-015, BE-017
- Scope:
  - добавить e2e integration-cases для настроек и шагов.
- Acceptance:
  - CI зеленый для новых и существующих шаговых endpoint.

## BE-019: Integration-тесты авторизации через `initData`
- Estimate: 1d
- Depends on: BE-004
- Scope:
  - добавить тесты на прием `initData`,
  - проверить выдачу `accessToken` и `refreshToken` в ответе.
- Acceptance:
  - payload с `initData` проходит успешно и выдает валидную token pair.

## BE-006: Auth sessions в БД + refresh rotation
- Estimate: 2d
- Depends on: BE-004
- Scope:
  - миграция `auth_sessions` (hash access/refresh, jti, revoke/rotate поля),
  - выпуск токенов с `jti`,
  - `POST /api/auth/refresh` через lookup сессии в БД,
  - rotation: старая refresh-сессия отзывается, создается новая.
- Acceptance:
  - refresh работает только для активной сессии,
  - повторное использование старого refresh возвращает `401`,
  - logout-all отзывает все активные сессии пользователя.

## Milestone M2: Полнота UX по БАД

## BE-020: Домен - нормализация статуса skipped->missed
- Estimate: 1d
- Depends on: BE-002
- Scope:
  - принимать `skipped` на входе, хранить/отдавать канонический `missed`.
- Acceptance:
  - compatibility-тесты на оба входных значения.

## BE-021: Medication entity/DTO с category + color
- Estimate: 1d
- Depends on: BE-011
- Scope:
  - протащить `category` и `color` через entity/repo/dto слои.
- Acceptance:
  - create/update/list включают оба поля.

## BE-022: API - POST /api/medication-logs/ensure-range
- Estimate: 2d
- Depends on: BE-020
- Scope:
  - материализовать pending-логи по расписанию на заданный диапазон,
  - обеспечить идемпотентность через unique `(medication_id, scheduled_at)`.
- Acceptance:
  - повторные вызовы не создают дубли.

## BE-023: API - GET /api/supplements/today
- Estimate: 1d
- Depends on: BE-021, BE-022
- Scope:
  - вернуть таблетки + спортпит за выбранный день,
  - включить маркер `isCompletedForDay`.
- Acceptance:
  - ответ содержит category-группировку и summary.

## BE-024: API - GET /api/supplements/calendar
- Estimate: 2d
- Depends on: BE-021, BE-022
- Scope:
  - вернуть month-grid с counters по категориям на каждый день.
- Acceptance:
  - полное покрытие месяца и корректная агрегация.

## BE-025: API - GET /api/supplements/day
- Estimate: 1d
- Depends on: BE-021, BE-022
- Scope:
  - day drill-down для past/future сценариев.
- Acceptance:
  - прошлые дни содержат taken/missed, будущие — planned записи.

## BE-026: Набор integration-тестов по БАД
- Estimate: 1d
- Depends on: BE-023, BE-024, BE-025
- Scope:
  - integration-сценарии для today/calendar/day endpoint.
- Acceptance:
  - тесты покрывают category split и status transitions.

## Milestone M3: Тренировки + шаблоны

## BE-030: Миграция - таблица workout_sets
- Estimate: 1d
- Depends on: BE-003
- Scope:
  - создать дочернюю таблицу с order/reps/weight,
  - добавить индексы и FK cascade.
- Acceptance:
  - миграция up/down проверена.

## BE-031: Миграция - таблицы training_templates
- Estimate: 1d
- Depends on: BE-003
- Scope:
  - создать `training_templates` и `training_template_items`.
- Acceptance:
  - миграция up/down проверена.

## BE-032: Repo - workouts с массивом подходов
- Estimate: 2d
- Depends on: BE-030
- Scope:
  - транзакционные create/update/read для workout + sets.
- Acceptance:
  - при сбое нет частичных записей.

## BE-033: API - расширить контракты workouts для sets[]
- Estimate: 1d
- Depends on: BE-032
- Scope:
  - обновить ATO/DTO/mappers/openapi для per-set payload.
- Acceptance:
  - GET/POST/PUT workouts отдают и принимают массив подходов.

## BE-034: Repo/usecase - CRUD шаблонов тренировок
- Estimate: 2d
- Depends on: BE-031
- Scope:
  - добавить методы repository и usecase для templates/items.
- Acceptance:
  - CRUD-операции проходят unit + integration тесты.

## BE-035: API - endpoint /api/training-templates
- Estimate: 1d
- Depends on: BE-034
- Scope:
  - добавить handlers list/create/update/delete.
- Acceptance:
  - endpoint доступны в OpenAPI и роутере.

## BE-036: API - GET /api/workouts/activity
- Estimate: 1d
- Depends on: BE-032
- Scope:
  - monthly activity calendar payload для вкладки истории.
- Acceptance:
  - возвращает все дни месяца с `completedTrainings`.

## BE-037: Метаданные упражнений + динамические категории
- Estimate: 1d
- Depends on: BE-001
- Scope:
  - включить `muscleGroup` и `emoji` в контракт упражнений,
  - отдать динамические категории на основе dataset упражнений.
- Acceptance:
  - категории фильтра каталога формируются по данным API.

## BE-038: Integration-тесты тренировок и шаблонов
- Estimate: 1d
- Depends on: BE-033, BE-035, BE-036, BE-037
- Scope:
  - end-to-end сценарии для создания шаблонов и истории тренировок.
- Acceptance:
  - CI покрывает полный training flow.

## Milestone M4: Дашборд + hardening

## BE-040: API - агрегирующий endpoint GET /api/dashboard
- Estimate: 2d
- Depends on: BE-017, BE-025, BE-036
- Scope:
  - агрегировать today summary по шагам, БАД, тренировкам и alerts.
- Acceptance:
  - один endpoint покрывает above-the-fold дашборда.

## BE-041: Оптимизация запросов и индексы
- Estimate: 1d
- Depends on: BE-024, BE-040
- Scope:
  - добавить недостающие индексы из требований,
  - оптимизировать тяжелые monthly-запросы.
- Acceptance:
  - query plans валидированы на staging-sized dataset.

## BE-042: Централизованные input-ограничения
- Estimate: 1d
- Depends on: BE-002
- Scope:
  - enforce глобальные guardrails (`limit<=200`, границы диапазона дат и т.д.).
- Acceptance:
  - невалидные ranges/limits стабильно возвращают `invalid_input`.

## BE-043: Завершить метрики и структурные логи
- Estimate: 1d
- Depends on: BE-040
- Scope:
  - метрики latency/error counters по route,
  - гарантировать request_id + user_id в error-логах.
- Acceptance:
  - observability dashboard показывает endpoint-level видимость.

## BE-044: Набор contract-тестов по OpenAPI
- Estimate: 1d
- Depends on: BE-001, BE-040
- Scope:
  - добавить schema-based validation тесты ответов.
- Acceptance:
  - все покрытые endpoint проходят contract checks в CI.

## BE-045: Регрессия и performance-gate
- Estimate: 1d
- Depends on: BE-041, BE-043, BE-044
- Scope:
  - прогнать regression-pack и p95 benchmark-gates.
- Acceptance:
  - p95 read <250ms, write <350ms на staging-профиле.

## Milestone M5: Релиз

## BE-050: Release runbook + rollout checklist
- Estimate: 1d
- Depends on: BE-045
- Scope:
  - финализировать порядок миграций, rollback steps, canary checklist.
- Acceptance:
  - runbook сохранен и reviewed командой.

## BE-051: Прод-выкатка и 72h стабилизация
- Estimate: 2d
- Depends on: BE-050
- Scope:
  - deploy, monitoring, triage hot issues, проверка совместимости.
- Acceptance:
  - нет нерешенных P1/P2 после 72 часов.

## Рекомендуемая параллелизация

1. Трек A (ядро данных/API):
- BE-006, BE-010..BE-018, BE-020..BE-026

2. Трек B (тренировки/шаблоны):
- BE-030..BE-038

3. Трек C (hardening/release):
- BE-040..BE-051

## Порядок приоритета (если ограничена емкость команды)

1. BE-001, BE-004, BE-003, BE-010, BE-012, BE-013, BE-015, BE-017
2. BE-006, BE-020, BE-021, BE-022, BE-023, BE-024, BE-025
3. BE-030, BE-032, BE-033, BE-034, BE-035
4. BE-040, BE-041, BE-044, BE-045
