# BE-007..BE-016 — Steps: полный вертикальный срез

## Метаданные
- ID: `BE-007..BE-016`
- Приоритет: `P0`
- Статус: `Todo`
- Домен: `steps`

## Цель
Реализовать `steps` как отдельный DDD-vertical slice: `domain -> application -> infra -> delivery -> wiring -> tests -> openapi`.

## API-контракт (целевой)
1. `GET /api/steps?from=YYYY-MM-DD&to=YYYY-MM-DD`
- Выход `200`: список дневных записей.
2. `PUT /api/steps/:date`
- Вход: `{ count, source }`
- Выход `200`: запись за дату после upsert.
3. `POST /api/steps/add`
- Вход: `{ date, delta, source }`
- Выход `200`: запись за дату после атомарного изменения.
4. `DELETE /api/steps/:date`
- Выход `204`.
5. `GET /api/steps/analytics?month=YYYY-MM`
- Выход `200`: week/month analytics + series.

## Декомпозиция задач

### BE-007 — Domain model `steps`
- Сделать:
1. `DailyEntry` с доменными инвариантами.
2. `Source` enum + parser.
3. Валидации `userID > 0`, `count >= 0`.
- Артефакты:
1. `/Users/eboshit/projects/tracker/backend/internal/domain/steps/entry.go`
- DoD:
1. Невалидные значения возвращают `shared.ErrInvalidInput`.
2. Нет зависимостей на `infra` и `delivery`.

### BE-008 — Domain analytics model
- Сделать:
1. `Point` и `Analytics`.
2. Валидации для метрик и series.
- Артефакты:
1. `/Users/eboshit/projects/tracker/backend/internal/domain/steps/analytics.go`
- DoD:
1. Нельзя сформировать analytics с отрицательными значениями.

### BE-009 — Domain ports
- Сделать:
1. `Repository` контракт: `GetByDateRange`, `UpsertByDate`, `AddDelta`, `DeleteByDate`.
2. Контракты зависимостей для analytics (goal/timezone через существующий `user` read port).
- Артефакты:
1. `/Users/eboshit/projects/tracker/backend/internal/domain/steps/ports.go`
- DoD:
1. Порты покрывают все use case из API-контракта.

### BE-010 — Application `ListSteps`
- Сделать:
1. `Input/Output`.
2. Валидацию диапазона `from/to`.
3. Чтение через repo.
- Артефакты:
1. `/Users/eboshit/projects/tracker/backend/internal/application/steps/list.go`
- DoD:
1. Нет SQL/HTTP-логики в use case.

### BE-011 — Application `PutSteps`
- Сделать:
1. `Input/Output`.
2. Валидацию `date/count/source`.
3. Выполнение через `WithinTx(ctx, fn)`.
- Артефакты:
1. `/Users/eboshit/projects/tracker/backend/internal/application/steps/put.go`
- DoD:
1. Мутация завернута в UoW.

### BE-012 — Application `AddStepsDelta`
- Сделать:
1. `Input/Output`.
2. Запрет `delta == 0`.
3. Транзакционное атомарное изменение через repo.
- Артефакты:
1. `/Users/eboshit/projects/tracker/backend/internal/application/steps/add.go`
- DoD:
1. Нельзя получить отрицательный итог шагов.

### BE-013 — Application `DeleteSteps`
- Сделать:
1. `Input`.
2. Удаление через `WithinTx`.
3. Единое поведение для not found (зафиксировать в коде и тестах).
- Артефакты:
1. `/Users/eboshit/projects/tracker/backend/internal/application/steps/delete.go`
- DoD:
1. Поведение idempotency/404 стабильно и покрыто тестами.

### BE-014 — Application `StepsAnalytics`
- Сделать:
1. `Input/Output`.
2. Расчет week/month по timezone пользователя.
3. Подсчет fact/goal/progress + series.
- Артефакты:
1. `/Users/eboshit/projects/tracker/backend/internal/application/steps/analytics.go`
- DoD:
1. Расчет корректен для timezone пользователя.

### BE-015 — Infra Postgres `steps_repository`
- Сделать:
1. Реализацию `steps.Repository`.
2. Вынести SQL в `const` в начале файла.
3. Корректный `scan` и mapping ошибок.
- Артефакты:
1. `/Users/eboshit/projects/tracker/backend/internal/infra/postgres/steps_repository.go`
- DoD:
1. SQL только в `infra`.
2. `pgx.ErrNoRows` и доменные ошибки маппятся последовательно.

### BE-016 — Delivery + wiring + tests + OpenAPI
- Сделать:
1. `stepsapi` handler (`ATO -> Input -> use case -> Output -> DTO`).
2. Подключение routes в `router.go` под JWT middleware.
3. Wiring в `cmd/api/main.go`.
4. Unit-тесты application.
5. Handler-тесты delivery.
6. Обновление `/Users/eboshit/projects/tracker/backend/openapi.yaml`.
- Артефакты:
1. `/Users/eboshit/projects/tracker/backend/internal/delivery/http/stepsapi/handler.go`
2. `/Users/eboshit/projects/tracker/backend/internal/delivery/http/stepsapi/ato.go`
3. `/Users/eboshit/projects/tracker/backend/internal/delivery/http/stepsapi/dto.go`
4. `/Users/eboshit/projects/tracker/backend/cmd/api/main.go`
5. `/Users/eboshit/projects/tracker/backend/internal/delivery/http/server/router.go`
6. `/Users/eboshit/projects/tracker/backend/openapi.yaml`
- DoD:
1. `go test ./...` зеленый.
2. Все 5 endpoints отвечают по контракту.
3. Ошибки в едином формате `{code,error}`.

## Порядок выполнения
1. `BE-007 -> BE-009` (domain + ports).
2. `BE-010 -> BE-014` (application).
3. `BE-015` (infra).
4. `BE-016` (delivery, wiring, tests, openapi).

## Ссылки
- План: `/Users/eboshit/projects/tracker/backend/IMPLEMENTATION_PLAN.md`
- Backlog: `/Users/eboshit/projects/tracker/backend/IMPLEMENTATION_TASKS.md`
- Требования: `/Users/eboshit/projects/tracker/backend/BACKEND_REQUIREMENTS.md`
