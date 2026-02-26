# Требования к Backend: Health Tracking Mini App

## 1. Назначение

Этот документ определяет продуктовые и API-требования к backend, необходимые для поддержки текущего Vite-фронтенда (`/frontend`) на production-уровне качества.

Документ покрывает:
- функциональные требования,
- доменные контракты (сущности и payload),
- список API-методов,
- валидации и бизнес-правила,
- нефункциональные требования (безопасность, производительность, наблюдаемость),
- расхождения между текущей backend-реализацией и целевым поведением фронтенда.

## 2. Область

Входит в scope:
- Telegram-аутентификация и профиль пользователя,
- учет шагов и управление целями,
- БАД: категории, расписания, дневные логи приема, календарь месяца,
- каталог упражнений, шаблоны тренировок, логи тренировок, история активности,
- дашборд и аналитика/статистика,
- backend-поведение, готовое для напоминаний.

Вне scope:
- платежи/подписки,
- социальные функции,
- интеграции с wearable-устройствами (могут быть добавлены позже).

## 3. Пользовательские сценарии, которые нужно поддержать

1. Пользователь открывает mini app, проходит аутентификацию через Telegram (`initData`), получает JWT-токены.
2. Пользователь видит сводку на сегодня:
   - прогресс шагов относительно дневной цели,
   - прогресс плана по шагам за 7 дней и текущий месяц,
   - дневную приверженность по БАД,
   - количество тренировок за день,
   - графики шагов за 7 дней и 1 месяц.
3. Пользователь добавляет шаги быстрыми кнопками (+500/+1K/+2K/+5K/+10K) или вручную.
4. Пользователь меняет дневную цель шагов.
5. Пользователь видит БАД по категориям:
   - таблетки,
   - спортпит.
6. Пользователь отмечает прием БАД как принятый/отклоненный.
7. Пользователь создает/обновляет/удаляет БАД и время приема.
8. Пользователь смотрит календарь БАД за текущий месяц и проваливается в конкретный день.
9. Пользователь просматривает каталог упражнений с динамическими категориями на основе списка упражнений.
10. Пользователь создает и редактирует шаблоны тренировок (из упражнений).
11. Пользователь выполняет тренировку, логирует подходы, сохраняет историю.
12. Пользователь смотрит историю тренировок и календарь активности за месяц.

## 4. Правила проектирования API

Базовый путь:
- `/api`

Аутентификация:
- Bearer JWT (`Authorization: Bearer <access_token>`) для защищенных маршрутов,
- refresh flow через endpoint обновления токена.
- логин/пароль не используются; вход выполняется только через Telegram WebApp payload.

Время и даты:
- поля только даты: `YYYY-MM-DD`,
- timestamp: RFC3339 UTC (`2026-02-24T10:30:00Z`),
- вычисления, показываемые пользователю, должны учитывать таймзону пользователя.

Единый формат ошибок:
```json
{
  "code": "invalid_input | unauthorized | not_found | conflict | internal_error",
  "error": "human readable message"
}
```

Пагинация и сортировка:
- query-параметры: `limit`, `offset`, `sortBy`, `sortDir`,
- значения по умолчанию: `limit=50`, `offset=0`,
- максимальный `limit`: `200`.

Идемпотентность:
- `PUT` операции должны быть идемпотентными,
- `POST` операции, которые могут ретраиться клиентом, рекомендуется поддерживать с заголовком `Idempotency-Key`.

## 5. Доменные контракты

## 5.1 Пользователь
```json
{
  "id": 1,
  "tgId": 123456789,
  "username": "john",
  "firstName": "John",
  "lastName": "Doe",
  "timezone": "Europe/Moscow",
  "stepsGoal": 10000,
  "streak": 5,
  "createdAt": "2026-02-22T10:00:00Z",
  "updatedAt": "2026-02-24T09:00:00Z"
}
```

Примечание:
- `stepsGoal` и `streak` обязательны для текущего поведения фронтенда.

## 5.2 Дневная запись шагов
```json
{
  "id": 10,
  "userId": 1,
  "date": "2026-02-24",
  "count": 7543,
  "source": "manual",
  "createdAt": "2026-02-24T06:00:00Z",
  "updatedAt": "2026-02-24T10:00:00Z"
}
```

## 5.2.1 Ответ аналитики шагов (7д / месяц)
```json
{
  "goalPerDay": 10000,
  "week": {
    "from": "2026-02-18",
    "to": "2026-02-24",
    "goalTotal": 70000,
    "factTotal": 54230,
    "completionPercent": 77.47,
    "series": [
      { "date": "2026-02-18", "steps": 7200 },
      { "date": "2026-02-19", "steps": 9400 }
    ]
  },
  "month": {
    "month": "2026-02",
    "goalTotal": 280000,
    "factTotal": 163450,
    "completionPercent": 58.38,
    "series": [
      { "date": "2026-02-01", "steps": 4500 },
      { "date": "2026-02-02", "steps": 8200 }
    ]
  }
}
```

## 5.3 БАД (Medication)
```json
{
  "id": 21,
  "userId": 1,
  "name": "Vitamin D3",
  "category": "tablet",
  "dose": 1000,
  "unit": "IU",
  "color": "lime",
  "schedule": {
    "byDay": ["MO", "TU", "WE", "TH", "FR", "SA", "SU"],
    "times": ["09:00"]
  },
  "active": true,
  "createdAt": "2026-02-22T10:00:00Z",
  "updatedAt": "2026-02-24T09:00:00Z"
}
```

Обязательные enum:
- `color`: `lime | violet | rose | info`
- `category`: `tablet | sport_nutrition`

## 5.4 Лог приема БАД
```json
{
  "id": 1001,
  "medicationId": 21,
  "userId": 1,
  "scheduledAt": "2026-02-24T09:00:00Z",
  "takenAt": "2026-02-24T09:05:00Z",
  "status": "taken",
  "createdAt": "2026-02-24T09:00:00Z",
  "updatedAt": "2026-02-24T09:05:00Z"
}
```

Обязательный enum:
- `status`: `pending | taken | missed`

Совместимость:
- backend может принимать `skipped` как alias и нормализовать в `missed`.

## 5.4.1 Детали дня по БАД (calendar drill-down)
```json
{
  "date": "2026-02-24",
  "isPast": true,
  "isFuture": false,
  "items": [
    {
      "medicationId": 21,
      "name": "Vitamin D3",
      "category": "tablet",
      "scheduledAt": "2026-02-24T09:00:00Z",
      "status": "missed"
    },
    {
      "medicationId": 22,
      "name": "Whey Protein",
      "category": "sport_nutrition",
      "scheduledAt": "2026-02-24T12:00:00Z",
      "status": "taken"
    }
  ],
  "summary": {
    "tablet": { "taken": 1, "missed": 1, "pending": 0 },
    "sport_nutrition": { "taken": 1, "missed": 0, "pending": 0 }
  }
}
```

## 5.5 Упражнение
```json
{
  "id": 31,
  "userId": 1,
  "name": "Squat",
  "muscleGroup": "Legs",
  "emoji": "🦵",
  "type": "strength",
  "durationSecDefault": 900,
  "notes": "",
  "createdAt": "2026-02-22T10:00:00Z",
  "updatedAt": "2026-02-24T09:00:00Z"
}
```

## 5.6 Тренировка + подходы
```json
{
  "id": 501,
  "userId": 1,
  "exerciseId": 31,
  "date": "2026-02-24",
  "startedAt": "2026-02-24T10:00:00Z",
  "endedAt": "2026-02-24T10:20:00Z",
  "durationSec": 1200,
  "sets": [
    { "order": 1, "reps": 12, "weightKg": 60 },
    { "order": 2, "reps": 10, "weightKg": 70 },
    { "order": 3, "reps": 8, "weightKg": 75 }
  ],
  "notes": "Felt good",
  "createdAt": "2026-02-24T10:20:00Z",
  "updatedAt": "2026-02-24T10:20:00Z"
}
```

Причина:
- фронтенд логирует детали по каждому подходу. Только агрегированных полей (`sets/reps/weight`) недостаточно.

## 5.8 Шаблон тренировки (новое)
```json
{
  "id": 901,
  "userId": 1,
  "name": "Leg Day A",
  "comment": "Base strength block",
  "items": [
    { "order": 1, "exerciseId": 31, "targetSets": 4, "targetReps": 8, "targetWeightKg": 70 },
    { "order": 2, "exerciseId": 32, "targetSets": 3, "targetReps": 10, "targetWeightKg": 24 }
  ],
  "createdAt": "2026-02-24T10:00:00Z",
  "updatedAt": "2026-02-24T10:30:00Z"
}
```

## 5.9 Календарь активности тренировок (новое)
```json
{
  "month": "2026-02",
  "days": [
    { "date": "2026-02-01", "completedTrainings": 0 },
    { "date": "2026-02-02", "completedTrainings": 1 }
  ]
}
```

## 5.10 Ответ дашборда
```json
{
  "date": "2026-02-24",
  "streak": 5,
  "steps": {
    "today": 7543,
    "goal": 10000,
    "progress": 0.7543,
    "remaining": 2457,
    "history7d": [
      { "date": "2026-02-18", "steps": 7200 },
      { "date": "2026-02-19", "steps": 9400 }
    ]
  },
  "medications": {
    "taken": 3,
    "total": 4,
    "missed": 1,
    "pending": 0
  },
  "workouts": {
    "todayCount": 1
  },
  "alerts": [
    { "type": "medication_missed", "medicationId": 21, "scheduledAt": "2026-02-24T09:00:00Z" }
  ]
}
```

## 6. Обязательные API-методы

## 6.1 Аутентификация и пользователь
- `POST /api/auth/telegram`
  - вход: `{ initData }`, опционально `X-Timezone`
  - выход: `{ user, token: {accessToken, refreshToken} }`
- `POST /api/auth/refresh`
  - вход: `{ refreshToken }`
  - выход: `{ accessToken, refreshToken }`
- `GET /api/me`
  - выход: полный профиль пользователя с `stepsGoal` и `streak`
- `PATCH /api/me/settings` (новое)
  - вход: `{ timezone?, stepsGoal? }`
  - выход: обновленный профиль

## 6.2 Шаги
- `GET /api/steps?from=YYYY-MM-DD&to=YYYY-MM-DD` (расширить текущий list endpoint диапазоном)
- `PUT /api/steps/{date}`
  - upsert точного значения на дату
  - вход: `{ count, source }`
- `POST /api/steps/add` (новый convenience endpoint)
  - вход: `{ date?, delta, source }`
  - атомарное инкрементирование для быстрых действий (`+500`, `+1000`, `+2000`, `+5000`, `+10000`)
- `GET /api/steps/analytics?month=YYYY-MM` (новое)
  - возвращает:
    - series графика за 7 дней,
    - series графика за месяц,
    - процент выполнения плана за неделю,
    - процент выполнения плана за месяц
- `DELETE /api/steps/{date}`

## 6.3 БАД и логи
- `GET /api/medications`
- `POST /api/medications`
- `PUT /api/medications/{id}`
- `DELETE /api/medications/{id}`
- `GET /api/medication-logs?from=RFC3339&to=RFC3339&medicationId&status&limit&offset`
- `POST /api/medication-logs`
  - ручное создание лога
- `PUT /api/medication-logs/{id}`
  - обновление статуса
- `POST /api/medication-logs/ensure-range` (новое)
  - вход: `{ from, to }`
  - гарантирует создание pending-логов для всех запланированных приемов в диапазоне
- `GET /api/supplements/today?date=YYYY-MM-DD` (новый convenience endpoint)
  - возвращает список всех таблеток и спортпита за выбранный день
- `GET /api/supplements/calendar?month=YYYY-MM` (новый convenience endpoint)
  - возвращает все дни месяца с дневными summary по категориям
- `GET /api/supplements/day?date=YYYY-MM-DD` (новый convenience endpoint)
  - возвращает детальный список по выбранному дню (поддержка прошлого/будущего)

## 6.4 Упражнения
- `GET /api/exercises?limit&offset&q&muscleGroup&sortBy&sortDir`
- `POST /api/exercises`
- `PUT /api/exercises/{id}`
- `DELETE /api/exercises/{id}`

## 6.5 Тренировки и сессии
- `GET /api/training-templates` (новое)
- `POST /api/training-templates` (новое)
- `PUT /api/training-templates/{id}` (новое)
- `DELETE /api/training-templates/{id}` (новое)
- `GET /api/workouts?from=RFC3339&to=RFC3339&exerciseId&limit&offset&sortBy&sortDir`
- `POST /api/workouts`
  - должен принимать массив подходов
- `PUT /api/workouts/{id}`
  - должен поддерживать замену списка подходов
- `DELETE /api/workouts/{id}`
- `GET /api/workouts/activity?month=YYYY-MM` (новое)
  - возвращает календарь активности для вкладки истории

## 6.6 Дашборд и статистика
- `GET /api/dashboard?date=YYYY-MM-DD` (новый агрегирующий endpoint)
- `GET /api/stats?range=day|week|month`
- `GET /api/stats?from=YYYY-MM-DD&to=YYYY-MM-DD`

## 7. Бизнес-правила и валидации

Общие:
- все защищенные endpoint должны жестко изолировать данные по `user_id`,
- `limit` в диапазоне `[1, 200]`, `offset >= 0`.

Шаги:
- `count >= 0`,
- `delta != 0` для add endpoint,
- диапазон цели: `1000..30000`, шаг `500`,
- должны поддерживаться пресеты быстрого добавления: `500`, `1000`, `2000`, `5000`, `10000`,
- недельный план:
  - `goalWeek = stepsGoal * 7`,
  - `completionWeek = factWeek / goalWeek`,
- месячный план:
  - `goalMonth = stepsGoal * daysInCurrentMonth`,
  - `completionMonth = factMonth / goalMonth`.

БАД:
- `name` не пустое после trim,
- `dose >= 0`,
- `schedule.times` должно быть валидным `HH:mm`,
- `schedule.byDay` должно быть подмножеством `MO..SU`,
- `category` должно быть `tablet` или `sport_nutrition`,
- при удалении БАД должен быть каскад или soft-delete будущих pending-логов (по продуктовому решению),
- в ответе "today" каждый элемент БАД должен включать `isCompletedForDay`, если все запланированные приемы за день имеют статус `taken`.

Логи БАД:
- допустимые переходы статусов:
  - `pending -> taken`,
  - `pending -> missed`,
  - `taken -> missed` (опционально, если нужны корректировки),
  - `missed -> taken` (опционально, если нужны корректировки),
- `takenAt` выставляется автоматически, когда статус становится `taken`,
- уникальность по `(medication_id, scheduled_at)` для идемпотентной генерации.

Тренировки:
- `endedAt >= startedAt`,
- `durationSec >= 0`,
- для каждого подхода: `reps > 0`, `weightKg >= 0`,
- минимум один подход обязателен,
- шаблон тренировки должен содержать минимум один элемент упражнения,
- категории в каталоге упражнений должны формироваться динамически из существующих данных (без хардкода на backend).

Статистика:
- если передан `from`/`to`, должны быть переданы оба,
- рекомендуемый максимальный диапазон дат: 366 дней,
- все вычисления по дням должны учитывать таймзону пользователя.

## 8. Требования к модели данных (БД)

Обязательные таблицы (или эквивалент):
- `users` (добавить `steps_goal`, опционально `streak_cache`),
- `steps` (unique `(user_id, date)`),
- `medications` (добавить `color`, `category`),
- `medication_logs` (unique `(medication_id, scheduled_at)`),
- `exercises` (добавить `muscle_group`, `emoji`),
- `workouts`,
- `workout_sets` (новая дочерняя таблица: `workout_id`, `order_no`, `reps`, `weight_kg`),
- `training_templates` (новая),
- `training_template_items` (новая, нормализованный список упражнений в шаблоне).

Индексы:
- `steps(user_id, date)`,
- `medication_logs(user_id, scheduled_at)`,
- `medication_logs(medication_id, scheduled_at)` unique,
- `workouts(user_id, started_at)`,
- `workout_sets(workout_id, order_no)`,
- `exercises(user_id, lower(name))`,
- `training_templates(user_id, updated_at)`,
- `training_template_items(template_id, order_no)`.

## 9. Нефункциональные требования

Цели по производительности:
- p95 для read endpoint < 250 ms (кроме auth endpoint),
- p95 для write endpoint < 350 ms.

Надежность:
- идемпотентное поведение при генерации расписаний и ретраях POST,
- транзакционная консистентность для записи workout + workout_sets.

Безопасность:
- JWT access + refresh с ротацией при refresh,
- обязательная верификация Telegram auth,
- CORS только для разрешенных origin,
- безопасное логирование (без утечек токенов).

Наблюдаемость:
- структурные логи с request id и user id,
- метрики: latency, error rate, DB query latency,
- health endpoint и проверки readiness зависимостей.

## 10. Требования к тестированию

Unit-тесты:
- нормализация и валидаторы домена,
- правила перехода статусов,
- расчеты диапазонов с учетом таймзон.

Integration-тесты:
- полный auth flow,
- CRUD-потоки для steps/medications/exercises/workouts,
- идемпотентность генерации medication schedule,
- корректность stats на фиксированном наборе данных.

Contract-тесты:
- валидация OpenAPI schema для всех endpoint,
- тесты обратной совместимости для alias статуса (`skipped` -> `missed`).

## 11. Расхождения с текущим Backend (важно)

Текущий backend уже имеет сильную базу (`/api/auth`, `/api/steps`, `/api/workouts`, `/api/medications`, `/api/medication-logs`, `/api/stats`), но для паритета с целевым UX нужны следующие доработки:

1. Добавить контракт настроек пользователя для `stepsGoal` (и, опционально, источник `streak`).
2. Добавить контракт аналитики шагов:
   - графики за неделю/месяц,
   - проценты выполнения плана за 7 дней и текущий месяц.
3. Добавить поля `color` и `category` для БАД.
4. Выровнять словарь статусов БАД с фронтендом (`missed` вместо текущего `skipped`).
5. Добавить convenience endpoint для БАД:
   - список на сегодня,
   - календарь месяца,
   - drill-down по дню.
6. Добавить display-поля упражнений (`muscleGroup`, `emoji`) или стабильный mapping metadata.
7. Поддержать хранение и API-контракты для подходов как массива (не только агрегаты).
8. Добавить CRUD шаблонов тренировок и endpoint календаря активности истории.
9. Добавить агрегирующий endpoint `/api/dashboard`.
10. Добавить материализацию логов в диапазоне для календаря (`ensure-range`) или эквивалентное серверное поведение.
11. Явно зафиксировать авторизацию через Telegram `initData` и немедленную выдачу `accessToken/refreshToken`.

## 12. Рекомендуемый план поставки

Фаза 1 (Core parity):
- авторизация через `initData` с немедленной выдачей access/refresh токенов,
- настройки пользователя (`stepsGoal`),
- недельно-месячная аналитика шагов,
- выравнивание статусов БАД,
- `color + category` для БАД,
- endpoint дашборда.

Фаза 2 (полнота UX по БАД):
- endpoint `today/calendar/day` для БАД,
- ensure-range для medication logs.

Фаза 3 (полнота по тренировкам):
- таблица workout sets + контракты,
- metadata упражнений,
- CRUD шаблонов тренировок,
- endpoint календаря активности тренировок.

Фаза 4 (стабилизация качества):
- idempotency keys,
- тюнинг производительности,
- полный набор контрактных тестов.
