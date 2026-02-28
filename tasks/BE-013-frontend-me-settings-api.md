# Me Settings API для Frontend

Базовый префикс: `/api`  
Авторизация: `Authorization: Bearer <accessToken>`

## Общие форматы

- `timezone`: IANA timezone (`Europe/Moscow`, `Asia/Almaty`, `America/New_York`).
- `stepsGoal`: `1000..30000` с шагом `500`.
- `createdAt`, `updatedAt`: UTC `RFC3339` (`2026-02-27T12:00:00Z`).

## Единый формат ошибок

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

## PATCH /api/me/settings

Partial update пользовательских настроек.

Можно передавать:
- только `timezone`,
- только `stepsGoal`,
- оба поля вместе.

Важно: пустой body `{}` или body без этих полей вернет `400`.

### Запрос: обновить только цель

`PATCH /api/me/settings`

```json
{
  "stepsGoal": 12000
}
```

### Запрос: обновить только timezone

`PATCH /api/me/settings`

```json
{
  "timezone": "Europe/Moscow"
}
```

### Запрос: обновить оба поля

`PATCH /api/me/settings`

```json
{
  "timezone": "Europe/Moscow",
  "stepsGoal": 12000
}
```

Ответ `200`:

```json
{
  "user": {
    "id": 1,
    "tgId": 123456789,
    "username": "john",
    "firstName": "John",
    "lastName": "Doe",
    "timezone": "Europe/Moscow",
    "stepsGoal": 12000,
    "streak": 3,
    "createdAt": "2026-02-22T10:00:00Z",
    "updatedAt": "2026-02-27T12:00:00Z"
  }
}
```

Нюансы:
- если `timezone` передан, он должен быть валидным IANA;
- если `stepsGoal` передан, должен проходить доменную валидацию (диапазон + шаг);
- при успешном обновлении фронт получает полный актуальный профиль пользователя.

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

export interface UserProfile {
  id: number;
  tgId: number;
  username: string;
  firstName: string;
  lastName?: string;
  timezone: string;
  stepsGoal: number;
  streak: number;
  createdAt: string; // RFC3339 UTC
  updatedAt: string; // RFC3339 UTC
}

export interface MeSettingsPatchRequest {
  timezone?: string;
  stepsGoal?: number;
}

export interface MeSettingsPatchResponse {
  user: UserProfile;
}
```
