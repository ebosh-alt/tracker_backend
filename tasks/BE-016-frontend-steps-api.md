# Steps API для Frontend

Базовый префикс: `/api`  
Авторизация для всех `steps` endpoint: `Authorization: Bearer <accessToken>`

## Общие форматы

- `date`: строка `YYYY-MM-DD` (локальная дата без времени).
- `month`: строка `YYYY-MM`.
- `createdAt`, `updatedAt`: UTC `RFC3339` (`2026-02-27T10:05:00Z`).
- `source`: сейчас поддерживается только `"manual"`.

## Единый формат ошибок

Все ошибки возвращаются как:

```json
{
  "code": "invalid_input",
  "error": "human readable message"
}
```

Возможные `code`:
- `invalid_input` (`400`)
- `unauthorized` (`401`)
- `not_found` (`404`)
- `conflict` (`409`)
- `internal_error` (`500`)

## 1) Получить записи шагов

`GET /api/steps?from=YYYY-MM-DD&to=YYYY-MM-DD`

Ответ `200`:

```json
{
  "entries": [
    {
      "date": "2026-02-27",
      "count": 1500,
      "source": "manual",
      "createdAt": "2026-02-27T10:00:00Z",
      "updatedAt": "2026-02-27T10:05:00Z"
    }
  ]
}
```

Нюансы:
- Диапазон включительный (`from..to`).
- Если записей нет, возвращается `entries: []`.
- `from` должен быть `<= to`, иначе `400`.

## 2) Upsert записи за дату

`PUT /api/steps/:date`

Тело:

```json
{
  "count": 1500,
  "source": "manual"
}
```

Ответ `200`:

```json
{
  "entry": {
    "date": "2026-02-27",
    "count": 1500,
    "source": "manual",
    "createdAt": "2026-02-27T10:00:00Z",
    "updatedAt": "2026-02-27T10:05:00Z"
  }
}
```

Нюансы:
- `date` берется из path-параметра.
- По бизнес-логике `count >= 0`.
- Текущее ограничение реализации: `count = 0` вернет `400` из-за `binding:"required"` для числового поля.

## 3) Добавить дельту к записи

`POST /api/steps/add`

Тело:

```json
{
  "date": "2026-02-27",
  "delta": 500,
  "source": "manual"
}
```

Ответ `200`:

```json
{
  "entry": {
    "date": "2026-02-27",
    "count": 2000,
    "source": "manual",
    "createdAt": "2026-02-27T10:00:00Z",
    "updatedAt": "2026-02-27T10:10:00Z"
  }
}
```

Нюансы:
- `delta` не может быть `0`.
- Отрицательная `delta` допустима, но итоговый `count` не может стать отрицательным (иначе `400`).

## 4) Удалить запись за дату

`DELETE /api/steps/:date`

Ответ `204` без body.

Нюансы:
- Операция идемпотентная: если записи нет, все равно `204`.

## 5) Аналитика за месяц + текущая неделя

`GET /api/steps/analytics?month=YYYY-MM`

Ответ `200`:

```json
{
  "goalPerDay": 10000,
  "week": {
    "from": "2026-02-21",
    "to": "2026-02-27",
    "goalTotal": 70000,
    "factTotal": 54230,
    "completionPercent": 77.47,
    "series": [
      { "date": "2026-02-24", "steps": 7200 }
    ]
  },
  "month": {
    "from": "2026-02-01",
    "to": "2026-02-28",
    "goalTotal": 280000,
    "factTotal": 163450,
    "completionPercent": 58.38,
    "series": [
      { "date": "2026-02-01", "steps": 4500 }
    ]
  }
}
```

Нюансы:
- Week считается как последние 7 дней (`to=today`) в timezone пользователя.
- Month считается по границам выбранного месяца в timezone пользователя.
- `series` содержит только существующие записи (без принудительного заполнения нулями).

## TypeScript модели (рекомендуемые)

```ts
export type ApiErrorCode =
  | "invalid_input"
  | "unauthorized"
  | "not_found"
  | "conflict"
  | "internal_error";

export interface ApiError {
  code: ApiErrorCode;
  error: string;
}

export interface DailyEntry {
  date: string; // YYYY-MM-DD
  count: number;
  source: "manual";
  createdAt: string; // RFC3339 UTC
  updatedAt: string; // RFC3339 UTC
}

export interface StepsListResponse {
  entries: DailyEntry[];
}

export interface StepsPutRequest {
  count: number;
  source: "manual";
}

export interface StepsPutResponse {
  entry: DailyEntry;
}

export interface StepsAddRequest {
  date: string; // YYYY-MM-DD
  delta: number; // != 0
  source: "manual";
}

export interface StepsAddResponse {
  entry: DailyEntry;
}

export interface AnalyticsPoint {
  date: string; // YYYY-MM-DD
  steps: number;
}

export interface AnalyticsPeriod {
  from: string; // YYYY-MM-DD
  to: string;   // YYYY-MM-DD
  goalTotal: number;
  factTotal: number;
  completionPercent: number;
  series: AnalyticsPoint[];
}

export interface StepsAnalyticsResponse {
  goalPerDay: number;
  week: AnalyticsPeriod;
  month: AnalyticsPeriod;
}
```
