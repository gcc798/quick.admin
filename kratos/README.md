# Kratos Rewrite Plan

## Goal

Build a brand-new backend under `kratos/` using the Kratos framework, while keeping:

- `native/` as the business baseline and read-only reference.
- `web/` working without frontend API changes.
- `sys-api` as the external HTTP service.
- `sys-rpc` as the internal gRPC service owning business logic and database access.

## Architecture

This rewrite uses a monorepo layout, while each service keeps the standard Kratos internal layering.

```text
kratos/
├── api/
│   └── system/v1/
├── app/
│   ├── sys-api/
│   │   ├── cmd/server
│   │   ├── configs
│   │   └── internal/{biz,data,server,service,conf}
│   └── sys-rpc/
│       ├── cmd/server
│       ├── configs
│       └── internal/{biz,data,server,service,conf}
├── pkg/
├── third_party/
├── Makefile
└── go.mod
```

## Service Boundaries

### sys-api

Responsibilities:

- Expose HTTP endpoints to `web/`
- Parse request parameters
- Adapt responses to native-compatible HTTP shape
- Handle JWT/token/cache related concerns
- Call `sys-rpc` through gRPC

Non-responsibilities:

- No direct database access
- No core system business logic persistence

### sys-rpc

Responsibilities:

- Expose internal gRPC APIs
- Own all system domain business logic
- Own PostgreSQL access
- Own Ent schema, repositories, and transactions

## API Contract Strategy

Use proto as the single source of truth.

- `api/system/v1/*.proto` defines both HTTP and gRPC contracts.
- HTTP paths, methods, request fields, and response fields must align with `native/`.
- `web/` should not require adjustments unless `web/` itself deviates from `native/`.

## ORM Strategy

Use `ent` with PostgreSQL.

- Ent schema lives under `application/sys-rpc/ent/schema`
- Data access repositories live under `application/sys-rpc/internal/data`
- Business orchestration lives under `application/sys-rpc/internal/biz`

## Initial Development Scope

Phase 1:

- Create `kratos/` base project structure
- Create `sys-api` and `sys-rpc` services in Kratos style
- Create shared proto layout
- Create Makefile and generation/build commands
- Ensure the whole project compiles

Phase 2:

- Add first proto contracts for health/auth
- Generate HTTP/gRPC code
- Keep a minimal runnable skeleton

Later phases:

- Auth / captcha / me / menu
- User / role / org
- Dict / config
- Login log / oper log / storage env / attachment
- Full compatibility pass against `native/` and `web/`

## Implementation Rules

1. Do not modify `native/`
2. Keep Kratos conventions inside each service
3. Prefer generated code where Kratos supports generation
4. Keep `sys-api` free of direct DB access
5. Keep contracts strong-typed, no generic JSON RPC envelope

## Current Target

The current target for this iteration is:

- write this plan into the repository
- scaffold the Kratos project and service skeleton
- add the minimal shared proto/config/build setup
- make the project compile successfully

## Current Scaffold Status

The repository now contains a minimal compilable Kratos skeleton with:

- `sys-api` exposing `GET /health`
- `sys-api` calling `sys-rpc` through gRPC for the health check
- `sys-rpc` serving the corresponding `Ping` RPC
- `ent` generation bootstrapped in `application/sys-rpc/ent/schema`

## Make Targets

Common commands:

- `make init`
- `make proto-all`
- `make ent`
- `make fmt`
- `make build-all`
- `make test`

## Reusable Package Plan

The following infrastructure has already been extracted into `pkg/` and can be reused by future microservices:

- `pkg/configx`
  - shared YAML config loading
  - shared duration parsing helpers
- `pkg/grpcx`
  - shared gRPC client dial logic
  - support for both direct endpoint mode and discovery mode
- `pkg/registryx`
  - shared etcd client creation
  - shared service registrar creation
  - shared service discovery creation
- `pkg/metrics`
  - shared Prometheus metric definitions and helpers

The following items are good candidates for future extraction into `pkg/`, but should remain in service-local code for now. They are intentionally documented here first instead of being extracted immediately.

### Priority 1

- `pkg/httpx`
  - purpose:
    - shared HTTP response encoder
    - shared HTTP error encoder
    - shared native-compatible page/data response shaping
  - current source locations:
    - `application/sys-api/internal/server/codec.go`

- `pkg/authx`
  - purpose:
    - current-user context helpers
    - token/header extraction
    - client IP and user-agent extraction
    - reusable auth metadata propagation helpers
  - current source locations:
    - `application/sys-api/internal/data/context.go`

### Priority 2

- `pkg/observabilityx`
  - purpose:
    - shared HTTP metrics middleware
    - shared slow SQL / slow Redis logging helpers
    - shared observability bootstrap utilities
  - current source locations:
    - `application/sys-api/internal/server/metrics_handler.go`
    - `application/sys-rpc/internal/data/db_observability.go`
    - `application/sys-rpc/internal/data/redis_observability.go`

- `pkg/wsx`
  - purpose:
    - reusable websocket hub
    - reusable connection registry
    - reusable heartbeat and broadcast primitives
  - current source locations:
    - `application/sys-api/internal/server/websocket_hub.go`
    - `application/sys-api/internal/server/websocket_handler.go`

### Priority 3

- `pkg/permx`
  - purpose:
    - reusable operation-to-permission mapping helpers
    - reusable permission middleware scaffolding
  - note:
    - only worth extracting after permission naming is stable across services
  - current source locations:
    - `application/sys-api/internal/server/middleware.go`
    - `application/sys-api/internal/server/permissions.go`

- `pkg/operlogx`
  - purpose:
    - reusable operation-log middleware
    - reusable request/response truncation and normalization helpers
  - note:
    - only worth extracting after log DTO and auditing conventions are stable across services
  - current source locations:
    - `application/sys-api/internal/server/operlog_middleware.go`

## Extraction Rules For Future Work

Only move code into `pkg/` when all of the following are true:

1. The logic is not tied to one service's business semantics
2. At least two services can reasonably reuse it
3. The abstraction is already stable enough to avoid repeated churn
4. The extracted package does not need to import service-local `biz` or `data` code

Keep service-local code in place when the logic still carries system-specific assumptions, permission naming assumptions, or DTO coupling that may not generalize well to future microservices.
