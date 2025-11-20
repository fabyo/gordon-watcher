## Quick orientation for AI coding agents

This repo is a Go-based, generic file watcher + enqueuer. Focus on making small, safe changes that respect existing interfaces and config conventions.

Key facts (big picture)
- Entry point: `cmd/watcher/main.go` — it wires configuration, logger, metrics, health, telemetry, storage, queue and the `watcher` component.
- Core runtime: `internal/watcher` (Watcher, WorkerPool, RateLimiter, Cleaner, StabilityChecker). This component watches paths, stabilizes files, computes SHA256, uses storage for idempotency/locks and publishes a `queue.Message`.
- Queue implementations: `internal/queue` (see `queue.go` for the `Queue` interface and `rabbitmq.go` for RabbitMQ implementation). There is also a NoOp queue used when queue is disabled.
- Storage implementations: `internal/storage` (see `storage.go` for `Storage` interface, `redis.go` for Redis, `memory.go` for fallback memory storage). Storage provides processed/enqueued/failed markers and distributed locking.

Where config and env come from
- Config loader: `internal/config/*` uses Viper. It reads `config.yaml` from these locations (in order): `/etc/gordon/watcher`, `$HOME/.gordon/watcher`, `./configs`, `.`.
- Environment variables are prefixed with `GORDON_WATCHER_` and many critical keys are explicitly bound in `internal/config/config.go` (see the `viper.BindEnv(...)` lines). Do not remove or change those binds unless you intentionally change how env overrides work.
- Example configs live in `configs/` (e.g. `config.yaml`, `config.nfe.yaml`). Defaults are set in `internal/config/defaults.go` via `SetDefaults`.
- Important note about durations: config stores durations as integers (int64) and the program converts them with `time.Duration(...)` in `main.go` and `internal/watcher`. When adding duration fields follow the same pattern and set defaults as `int64(<duration>)` in `SetDefaults`.

Build / test / run workflows (use these exact targets)
- Build: `make build` → binary at `bin/gordon-watcher` (LDFLAGS embed version/commit/date).
- Run locally: `make run` (build then run). For quick dev hot-reload use `make dev` (requires `air`).
- Docker: `make dev-docker` runs `docker-compose up --build`; `make docker-build` builds the image.
- Unit tests: `make test` (runs `go test -v -race -timeout 30s ./...`).
- Integration tests: `make test-integration` (this runs tests under `tests/integration` with the `integration` build tag). Use `go test -tags=integration ./tests/integration/...` when running manually.
- Lint/format: `make lint` (uses `golangci-lint` if installed), `make fmt`, `make vet`, `make check`.

Project-specific patterns and conventions
- Config-binding is explicit: many env keys are `viper.BindEnv(...)`. If you add new env-overridable config keys, add matching `BindEnv` lines.
- Durations in config are stored as `int64` (nanoseconds) and converted with `time.Duration(...)` where used. See `cfg.Watcher.StableDelay` and `cfg.Watcher.CleanupInterval` usage in `main.go` and `internal/watcher`.
- Use the `moveFile` helper in `internal/watcher` to move files robustly (handles cross-device moves via copy+delete). Prefer it over naive `os.Rename` when moving files between devices.
- Idempotency: files are identified by SHA256 hash (`calculateHash`) and checked via `Storage.IsProcessed`. Respect that flow when implementing new processing steps.
- Distributed locking: storage provides `GetLock` returning a `Lock` with `Release(ctx)` — used to avoid double-processing. Follow the pattern in `watcher.processFile`.
- Logging: `internal/logger` provides a structured logger where callsites pass key/value pairs (e.g., `log.Info("msg", "key", val)`). Maintain key/value logging style for consistency.

Interfaces to follow when adding components
- `internal/queue/queue.go` — implement `Queue` (Publish, Close) and reuse `queue.Message` shape.
- `internal/storage/storage.go` — implement `Storage` methods: `IsProcessed`, `MarkEnqueued`, `MarkProcessed`, `MarkFailed`, `GetLock`, `Close`. The Redis implementation is a working example.

Examples & code pointers (replace or reference these when needed)
- Watcher wiring: `cmd/watcher/main.go` — shows how cfg → storage/queue → watcher are constructed and how health/metrics/telemetry are started.
- File processing flow: `internal/watcher/processFile`, `moveToProcessing`, `MarkEnqueued`, publish `queue.Message` → on publish failure move to failed and mark via storage.
- RabbitMQ example: `internal/queue/rabbitmq.go` (see exchange declare, queue declare, bind, `PublishWithContext`).
- Redis example: `internal/storage/redis.go` (keys/prefixes, lock TTL and Lua-based release).

Common developer tasks (examples)
- Enable queue and point to RabbitMQ:
  GORDON_WATCHER_QUEUE_ENABLED=true\
  GORDON_WATCHER_QUEUE_RABBITMQ_URL=amqp://guest:guest@localhost:5672/ \
  make run
- Run integration tests (if local infra needed, start docker compose then):
  docker-compose up -d rabbitmq redis
  make test-integration

Where to look first when debugging
- Config loading & env: `internal/config/config.go` and `internal/config/defaults.go` (missing binds cause surprising behavior).
- Watcher logic: `internal/watcher/watcher.go` (events, stability, rate limiting, worker pool). Follow `processFile` for the lifecycle of a detected file.
- Queue/storage failures: check `internal/queue/rabbitmq.go` and `internal/storage/redis.go` (connection errors cause fallback to NoOp or Memory storage in `main.go`).

Tests and examples
- Unit tests: `tests/unit` (see `pool_test.go`, `ratelimit_test.go`, `watcher_test.go`). Use `make test`.
- Integration: `tests/integration` and `Makefile` target `test-integration` (the suite is gated by `integration` tag).
- Example consumers / configs: `examples/` and `configs/` (use them as reference configs for behavior tuning).

Important DOs and DON'Ts for code changes
- DO: keep explicit viper.BindEnv lines for env-sensitive keys; add new binds when introducing important env overrides.
- DO: follow `Queue` and `Storage` interfaces exactly; prefer adding a `NewXxx(...)` constructor mirroring `NewRabbitMQQueue` / `NewRedisStorage` patterns.
- DO: preserve the file movement semantics (use `moveFile`) and idempotency/lock checks when changing processing flow.
- DON'T: assume durations in config are already time.Duration — they are stored as `int64` and converted at runtime.
- DON'T: remove debug prints in `internal/config/config.go` until you understand environment vs config precedence (they were added to help debug env overrides).

If anything in this file is unclear or you want more examples (e.g., a small PR template or specific unit-test scaffold), tell me which area and I will expand with snippets or a test harness.

---
Files referenced: `cmd/watcher/main.go`, `internal/config/*`, `internal/watcher/*`, `internal/queue/*`, `internal/storage/*`, `Makefile`, `configs/`, `tests/`.
