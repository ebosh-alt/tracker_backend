# Backend Architecture

## Target structure

```text
internal/
  domain/
    <context>/
      aggregate.go
      repository.go
  application/
    <context>/
      <use_case>.go
    uow.go
  infra/
    postgres/
      <repo>.go
      uow.go
  delivery/
    http/
      <context>api/
```

## Layer responsibilities

- `domain`:
  - entities/aggregates/value objects,
  - domain invariants and behavior,
  - repository interfaces declared from domain perspective.
- `application`:
  - one file per use case (`create.go`, `delete.go`, etc.),
  - orchestration of domain + repositories,
  - transaction boundaries through `UnitOfWork`.
- `infra`:
  - repository implementations (`Postgres`, external services),
  - `UnitOfWork` implementation (`WithinTx`),
  - no business rules.
- `delivery`:
  - transport concerns (HTTP/Gin),
  - ATO -> `application.Input`,
  - `application.Output` -> DTO/JSON,
  - error mapping to HTTP statuses.

## HTTP handler rules

Handler must do only:

1. parse request and validate request format,
2. extract `user_id` from auth middleware,
3. map request to `application.Input`,
4. call use case,
5. map `application.Output` to JSON DTO,
6. translate errors into HTTP codes.

Handler must not do:

- transactions,
- SQL,
- domain/business validations.
