**Следующий пул API (рекомендую сделать первым)**
1. `POST /api/auth/refresh`
2. `GET /api/me`
3. `PATCH /api/me/settings`

Этот пул логично следующий после уже сделанного `POST /api/auth/telegram`: закрывает полный auth/profile цикл и дает стабильную основу для остальных доменов.

**Точечный план реализации в новой архитектуре (DDD + Clean)**
1. `domain` (правила и контракты):
- Довести модель `user` как источник бизнес-правил настроек: timezone, stepsGoal, streak.
- Зафиксировать инварианты в [user.go](/Users/eboshit/projects/tracker/backend/internal/domain/user/user.go).
- Контракты портов пользователя держать в [ports.go](/Users/eboshit/projects/tracker/backend/internal/domain/user/ports.go), без SQL/HTTP деталей.

2. `application` (use cases):
- Добавить `internal/application/user/get_me.go` (чтение профиля по `UserID`).
- Добавить `internal/application/user/update_settings.go` (partial update, валидация через домен, `WithinTx`).
- Добавить `internal/application/auth/refresh.go` (нормализация refresh token, parse, проверка пользователя, выпуск новой пары токенов).

3. `infra` (адаптеры/репозитории):
- Реализовать недостающие методы в [user_repo.go](/Users/eboshit/projects/tracker/backend/internal/infra/postgres/user_repo.go): `GetByTelegramID`, `Save`, `UpdateSettings`.
- Все SQL вынести в `const` в начале файла (как ты и просил ранее).
- Для refresh use case добавить adapter к [jwt.go](/Users/eboshit/projects/tracker/backend/internal/infra/auth/jwt.go): `Parse(refresh)` + `Issue(userID)`.

4. `delivery/http` (тонкие handlers):
- Расширить [authapi/handler.go](/Users/eboshit/projects/tracker/backend/internal/delivery/http/authapi/handler.go): добавить `POST /auth/refresh`.
- Создать `meapi` пакет: `ato.go`, `dto.go`, `handler.go` для `GET /me` и `PATCH /me/settings`.
- В [router.go](/Users/eboshit/projects/tracker/backend/internal/delivery/http/server/router.go):
    - `POST /api/auth/refresh` в public group (без JWT),
    - `/api/me*` в protected group (с JWT).

5. `composition` (main wiring):
- Обновить [main.go](/Users/eboshit/projects/tracker/backend/cmd/api/main.go): собрать новые use case и передать в handlers.
- Сохранить текущий стиль: handler получает агрегат зависимостей (`UseCases struct`).

6. `tests` (обязательно по слоям):
- Unit: use case `refresh`, `get_me`, `update_settings`.
- Handler tests: маппинг ошибок (`invalid_input`, `unauthorized`, `not_found`, `internal_error`).
- Repo integration tests: `UpdateSettings` и readback.
- Финально: `go test ./...` + smoke сценарий auth->refresh->me->me/settings->me.

**Порядок выполнения (чтобы было проще)**
1. PR-1: `application` + `domain` (без HTTP).
2. PR-2: `infra/postgres` + JWT adapter.
3. PR-3: `delivery/http` + `router` + `main`.
4. PR-4: тесты и полировка ошибок/DTO.

**Критерий готовности пула**
1. Полный flow работает: `auth/telegram -> auth/refresh -> me -> me/settings`.
2. Handler’ы остаются тонкими (без SQL/транзакций/доменных правил).
3. Вся бизнес-валидация проходит через domain/application.
4. Все новые use case и handler покрыты тестами.

Если хочешь, следующим шагом разложу это в отдельный task-файл с ID (например `backend/tasks/BE-005-auth-refresh-and-me.md`) в формате, как у `BE-004`.