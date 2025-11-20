# Gordon Watcher - Project Structure
```
gordon-watcher/
├── cmd/
│   └── watcher/
│       └── main.go                    # Entry point
├── internal/
│   ├── config/                        # Configuration management
│   │   ├── config.go
│   │   ├── validator.go
│   │   └── defaults.go
│   ├── watcher/                       # Core watcher logic
│   │   ├── watcher.go
│   │   ├── pool.go                    # Worker pool
│   │   ├── ratelimit.go               # Rate limiting
│   │   ├── circuitbreaker.go          # Circuit breaker
│   │   ├── retry.go                   # Retry logic
│   │   ├── cleaner.go                 # Directory cleanup
│   │   └── stability.go               # File stability check
│   ├── queue/                         # Message queue
│   │   ├── queue.go
│   │   ├── message.go
│   │   ├── rabbitmq.go
│   │   └── noop.go
│   ├── storage/                       # State storage
│   │   ├── storage.go
│   │   ├── redis.go
│   │   └── memory.go
│   ├── health/                        # Health checks
│   │   └── health.go
│   ├── metrics/                       # Prometheus metrics
│   │   ├── prometheus.go
│   │   └── server.go
│   ├── telemetry/                     # OpenTelemetry
│   │   ├── tracer.go
│   │   └── span.go
│   └── logger/                        # Structured logging
│       └── logger.go
├── pkg/                               # Public packages
│   ├── filesystem/
│   │   ├── utils.go
│   │   └── hash.go
│   └── patterns/
│       └── matcher.go
├── configs/                           # Configuration files
│   ├── config.example.yaml
│   ├── config.dev.yaml
│   ├── config.nfe.yaml                # NF-e example
│   └── config.banking.yaml            # Banking example
├── ansible/                           # Deployment automation
│   ├── playbook.yml
│   ├── inventory/
│   ├── roles/
│   └── scripts/
├── docker/                            # Docker configs
│   ├── Dockerfile
│   ├── Dockerfile.dev
│   └── docker-compose.yml
├── scripts/                           # Helper scripts
│   ├── stress-test.sh
│   ├── generate-test-files.sh
│   └── benchmark.sh
├── tests/                             # Tests
│   ├── unit/
│   └── integration/
├── docs/                              # Documentation
│   ├── ARCHITECTURE.md
│   ├── CONFIGURATION.md
│   └── DEPLOYMENT.md
├── examples/                          # Usage examples
│   ├── nfe-processing/
│   ├── banking/
│   └── generic/
├── Makefile                           # Build automation
├── go.mod                             # Go dependencies
└── README.md                          # Project overview
```
