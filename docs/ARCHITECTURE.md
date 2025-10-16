# Architecture guide — Clean Architecture for `core`

This document recommends a folder structure and conventions to follow the Clean Architecture principles (separation of concerns, dependency rule) while keeping compatibility with the existing codebase.

Top-level layout

- `cmd/` — application entrypoints (if the repo will be used as an app). Keep small `main` packages here.
- `internal/` — packages only for this module (not importable by others). Use for application/business logic that shouldn't be public.
- `pkg/` — optional shared libraries intended for external reuse.
- `api/` — transport layer code (HTTP handlers, GRPC, CLI adapters).
- `config/` or `configx/` — configuration loader and related types.
  - Note: this repo provides a new `configx` package under `./configx` that contains a typed Loader implementing defaults (via `creasty/defaults`), decoding (mapstructure) and validation (go-playground/validator). Consider moving `configx` to `internal/config` if it's only used internally.
- `domain/` — core business entities, interfaces (usecases/repositories contracts).
- `usecase/` or `app/` — application use-cases / interactors implementing business rules.
- `infra/` — infrastructure implementations (DB, external clients, file storage).
- `docs/` — architecture docs, decision records, runbooks.
- `test/` — integration tests, test fixtures, mocks (optional)

Mapping of current files

- `app.go` — move to `cmd/app/main.go` (or keep at top-level if small). Should only wire dependencies and start the application.
- `config.go` — original helper kept in package `core`; we moved new loader to `configx/`. Consider moving to `config/` under `internal/config` if only used internally.
- `health.go` — belongs to `api/` or `internal/health` depending on whether it's an HTTP handler or internal check.
- `log.go` — infrastructure: move to `infra/logging` or `internal/logging`.
- `errors.go` — domain-level errors live in `domain/` or `pkg/errors` if shared.
- `core_test.go` — keep under the package it's testing; larger integration tests can go to `test/`.

Guidelines

- Keep packages small and focused. Prefer `internal/` for implementation that shouldn't be public.
- Define interfaces in the package that needs them (e.g., domain defines repository interfaces; infra implements them). This keeps the dependency direction from outer layers inward.
- Keep wiring (concrete implementations, fx, DI) at the application entrypoint (`cmd/`) or a dedicated `app/` package.
- Use clear names and avoid deep nesting. Typical layout:

```
cmd/
  app/
    main.go
internal/
  config/
    loader.go
  api/
    http/
      handlers.go
  infra/
    db/
    logging/
domain/
  model/
  ports/
usecase/
  user/
    service.go
docs/
  ARCHITECTURE.md
```

Migration plan (small incremental steps)

1. Add `internal/config` (or keep `configx`) and move loader there. Update imports.

Tests

- Unit tests for `configx` are included at `configx/config_test.go`. Run tests with:

```bash
GOWORK=off go test ./...
```

These tests cover a happy-path Bind and a validation failure path; they should pass before and after incremental refactors.
2. Create `domain` and move domain types/errors.
3. Create `infra` for logging and database clients.
4. Introduce `usecase` layer and move business logic.
5. Keep existing tests passing and add integration tests under `test/`.

Notes

- Keep go module tidy: run `go mod tidy` after moves.
- If you use `go.work`, update it to include module replacements during large refactors.
- I can provide a PR that moves files and updates imports incrementally to avoid a huge change.
