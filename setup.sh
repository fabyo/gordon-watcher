#!/bin/bash
# ═══════════════════════════════════════════════════════════
#  GORDON WATCHER - PROJECT SETUP SCRIPT
#  Creates entire project structure with empty files
# ═══════════════════════════════════════════════════════════

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

PROJECT_NAME="gordon-watcher"
GITHUB_USER="${GITHUB_USER:-fabyo}"

echo -e "${BLUE}╔═══════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║       GORDON WATCHER - PROJECT SETUP             ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════╝${NC}"
echo ""

# Check if already in project directory
if [ -f "go.mod" ] && grep -q "gordon-watcher" go.mod 2>/dev/null; then
    echo -e "${YELLOW}⚠ Already in gordon-watcher directory${NC}"
    read -p "Continue and overwrite? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Setup cancelled."
        exit 0
    fi
    PROJECT_DIR="."
else
    # Create project directory
    read -p "Project directory name [$PROJECT_NAME]: " input
    PROJECT_DIR="${input:-$PROJECT_NAME}"
    
    if [ -d "$PROJECT_DIR" ]; then
        echo -e "${RED}✗ Directory $PROJECT_DIR already exists${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}Creating project directory: $PROJECT_DIR${NC}"
    mkdir -p "$PROJECT_DIR"
    cd "$PROJECT_DIR"
fi

echo ""
echo -e "${YELLOW}📁 Creating directory structure...${NC}"

# ═══════════════════════════════════════════════════════════
#  CREATE DIRECTORIES
# ═══════════════════════════════════════════════════════════

directories=(
    # Main
    "cmd/watcher"
    
    # Internal
    "internal/config"
    "internal/watcher"
    "internal/queue"
    "internal/storage"
    "internal/health"
    "internal/metrics"
    "internal/telemetry"
    "internal/logger"
    
    # Pkg
    "pkg/filesystem"
    "pkg/patterns"
    
    # Configs
    "configs"
    
    # Ansible
    "ansible/inventory"
    "ansible/roles/gordon-watcher/tasks"
    "ansible/roles/gordon-watcher/templates"
    "ansible/roles/gordon-watcher/handlers"
    "ansible/roles/gordon-watcher/files"
    "ansible/roles/gordon-watcher/vars"
    "ansible/roles/gordon-watcher/defaults"
    "ansible/group_vars"
    "ansible/host_vars"
    "ansible/scripts"
    
    # Docker
    "docker/prometheus"
    "docker/grafana/dashboards"
    "docker/grafana/datasources"
    
    # Scripts
    "scripts"
    
    # Tests
    "tests/unit"
    "tests/integration"
    "tests/fixtures"
    
    # Examples
    "examples/banking"
    "examples/generic"
    
    # Docs
    "docs"
    
    # GitHub
    ".github/workflows"
    ".github/ISSUE_TEMPLATE"
    
    # Data directories (for development)
    "data/incoming"
    "data/processing"
    "data/processed"
    "data/failed"
    "data/ignored"
    "data/tmp"
    
    # Build
    "bin"
    "build"
    "coverage"
    "tmp"
)

for dir in "${directories[@]}"; do
    mkdir -p "$dir"
    echo -e "  ${GREEN}✓${NC} $dir"
done

echo ""
echo -e "${YELLOW}📝 Creating files...${NC}"

# ═══════════════════════════════════════════════════════════
#  CREATE EMPTY FILES
# ═══════════════════════════════════════════════════════════

files=(
    # Root
    "README.md"
    "LICENSE"
    "Makefile"
    ".gitignore"
    ".air.toml"
    "go.mod"
    "go.sum"
    ".env.example"
    "docker-compose.yml"
    
    # Cmd
    "cmd/watcher/main.go"
    
    # Internal - Config
    "internal/config/config.go"
    "internal/config/validator.go"
    "internal/config/defaults.go"
    
    # Internal - Watcher
    "internal/watcher/watcher.go"
    "internal/watcher/pool.go"
    "internal/watcher/ratelimit.go"
    "internal/watcher/circuitbreaker.go"
    "internal/watcher/retry.go"
    "internal/watcher/cleaner.go"
    "internal/watcher/stability.go"
    
    # Internal - Queue
    "internal/queue/queue.go"
    "internal/queue/message.go"
    "internal/queue/rabbitmq.go"
    "internal/queue/noop.go"
    
    # Internal - Storage
    "internal/storage/storage.go"
    "internal/storage/redis.go"
    "internal/storage/memory.go"
    
    # Internal - Health
    "internal/health/health.go"
    
    # Internal - Metrics
    "internal/metrics/prometheus.go"
    "internal/metrics/server.go"
    
    # Internal - Telemetry
    "internal/telemetry/tracer.go"
    "internal/telemetry/span.go"
    
    # Internal - Logger
    "internal/logger/logger.go"
    
    # Pkg
    "pkg/filesystem/utils.go"
    "pkg/filesystem/hash.go"
    "pkg/patterns/matcher.go"
    
    # Configs
    "configs/config.example.yaml"
    "configs/config.dev.yaml"
    "configs/config.staging.yaml"
    "configs/config.prod.yaml"
    "configs/config.prod.yaml"
    "configs/config.banking.yaml"
    
    # Ansible - Main
    "ansible/playbook.yml"
    "ansible/deploy.yml"
    "ansible/rollback.yml"
    "ansible/ansible.cfg"
    
    # Ansible - Inventory
    "ansible/inventory/development.yml"
    "ansible/inventory/staging.yml"
    "ansible/inventory/production.yml"
    
    # Ansible - Role Tasks
    "ansible/roles/gordon-watcher/tasks/main.yml"
    "ansible/roles/gordon-watcher/tasks/install.yml"
    "ansible/roles/gordon-watcher/tasks/configure.yml"
    "ansible/roles/gordon-watcher/tasks/systemd.yml"
    "ansible/roles/gordon-watcher/tasks/verify.yml"
    
    # Ansible - Role Templates
    "ansible/roles/gordon-watcher/templates/config.yaml.j2"
    "ansible/roles/gordon-watcher/templates/systemd.service.j2"
    "ansible/roles/gordon-watcher/templates/environment.j2"
    "ansible/roles/gordon-watcher/templates/logrotate.j2"
    
    # Ansible - Role Handlers
    "ansible/roles/gordon-watcher/handlers/main.yml"
    
    # Ansible - Role Vars
    "ansible/roles/gordon-watcher/vars/main.yml"
    "ansible/roles/gordon-watcher/defaults/main.yml"
    
    # Ansible - Group Vars
    "ansible/group_vars/all.yml"
    "ansible/group_vars/development.yml"
    "ansible/group_vars/staging.yml"
    "ansible/group_vars/production.yml"
    
    # Ansible - Scripts
    "ansible/scripts/deploy.sh"
    "ansible/scripts/rollback.sh"
    "ansible/scripts/healthcheck.sh"
    
    # Docker
    "docker/Dockerfile"
    "docker/Dockerfile.dev"
    "docker/prometheus/prometheus.yml"
    "docker/grafana/datasources/datasources.yml"
    
    # Scripts
    "scripts/build.sh"
    "scripts/test.sh"
    "scripts/deploy.sh"
    "scripts/stress-test.sh"
    "scripts/generate-test-files.sh"
    "scripts/benchmark.sh"
    "scripts/healthcheck.sh"
    
    # Tests
    "tests/unit/watcher_test.go"
    "tests/unit/pool_test.go"
    "tests/unit/ratelimit_test.go"
    "tests/integration/integration_test.go"
    
    # Examples
    "examples/banking/README.md"
    "examples/generic/README.md"
    
    # Docs
    "docs/ARCHITECTURE.md"
    "docs/CONFIGURATION.md"
    "docs/DEPLOYMENT.md"
    "docs/DEVELOPMENT.md"
    "docs/TROUBLESHOOTING.md"
    
    # GitHub
    ".github/workflows/test.yml"
    ".github/workflows/build.yml"
    ".github/workflows/release.yml"
    ".github/ISSUE_TEMPLATE/bug_report.md"
    ".github/ISSUE_TEMPLATE/feature_request.md"
    ".github/PULL_REQUEST_TEMPLATE.md"
    ".github/CODE_OF_CONDUCT.md"
    ".github/CONTRIBUTING.md"
)

for file in "${files[@]}"; do
    touch "$file"
    echo -e "  ${GREEN}✓${NC} $file"
done

echo ""
echo -e "${YELLOW}🔧 Making scripts executable...${NC}"

# Make scripts executable
chmod +x scripts/*.sh 2>/dev/null || true
chmod +x ansible/scripts/*.sh 2>/dev/null || true

# ═══════════════════════════════════════════════════════════
#  CREATE BASIC FILE CONTENTS
# ═══════════════════════════════════════════════════════════

echo ""
echo -e "${YELLOW}📋 Creating basic file contents...${NC}"

# .gitignore
cat > .gitignore << 'EOF'
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
gordon-watcher
dist/
bin/
build/

# Test binary
*.test

# Output of the go coverage tool
*.out
coverage/

# Dependency directories
vendor/

# Go workspace file
go.work

# IDEs
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Temporary files
tmp/
*.tmp
*.log

# Environment
.env
.env.local

# Data directories (development)
data/incoming/*
data/processing/*
data/processed/*
data/failed/*
data/ignored/*
data/tmp/*

# Keep directory structure
!data/incoming/.gitkeep
!data/processing/.gitkeep
!data/processed/.gitkeep
!data/failed/.gitkeep
!data/ignored/.gitkeep
!data/tmp/.gitkeep

# Ansible
ansible/*.retry
ansible/.vault_pass

# Coverage
coverage.html
EOF

echo -e "  ${GREEN}✓${NC} .gitignore"

# .gitkeep for data directories
for dir in incoming processing processed failed ignored tmp; do
    touch "data/$dir/.gitkeep"
done
echo -e "  ${GREEN}✓${NC} .gitkeep files"

# .env.example
cat > .env.example << EOF
# ═══════════════════════════════════════════════════════════
#  GORDON WATCHER - ENVIRONMENT VARIABLES
# ═══════════════════════════════════════════════════════════

# Application
GORDON_WATCHER_ENV=development
GORDON_WATCHER_LOG_LEVEL=debug

# Paths
GORDON_WATCHER_WORKING_DIR=/opt/gordon/data
GORDON_WATCHER_PATHS=/opt/gordon/data/incoming

# Workers
GORDON_WATCHER_MAX_WORKERS=10
GORDON_WATCHER_MAX_FILES_PER_SECOND=100

# Queue (RabbitMQ)
GORDON_WATCHER_QUEUE_ENABLED=true
GORDON_WATCHER_RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# Redis
GORDON_WATCHER_REDIS_ENABLED=false
GORDON_WATCHER_REDIS_ADDR=localhost:6379
GORDON_WATCHER_REDIS_PASSWORD=

# Telemetry
GORDON_WATCHER_TELEMETRY_ENABLED=true
GORDON_WATCHER_TELEMETRY_ENDPOINT=http://localhost:14268/api/traces

# Ports
GORDON_WATCHER_HEALTH_PORT=8081
GORDON_WATCHER_METRICS_PORT=9100
EOF

echo -e "  ${GREEN}✓${NC} .env.example"

# README.md skeleton
cat > README.md << EOF
# 🔨 Gordon Watcher

A powerful, generic file system watcher with extensible processing capabilities, inspired by Laravel's Artisan.

## Features

- 🎯 **Generic & Extensible** - Not tied to any specific use case
- 🔄 **Recursive Watching** - Monitors subdirectories automatically
- ⚡ **Worker Pool** - Prevents memory overflow with controlled concurrency
- 🚦 **Rate Limiting** - Protects downstream services
- 🔌 **Circuit Breaker** - Graceful degradation when services fail
- 🔁 **Retry Logic** - Exponential backoff for transient failures
- 📊 **Prometheus Metrics** - Full observability
- 🔍 **Distributed Tracing** - OpenTelemetry/Jaeger integration
- 🏥 **Health Checks** - Kubernetes-ready endpoints
- 🐳 **Docker Ready** - Complete containerization support
- 🤖 **Ansible Deployment** - Production-ready automation

## Quick Start

\`\`\`bash
# Build
make build

# Run
make run

# Development with hot reload
make dev
\`\`\`

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Configuration](docs/CONFIGURATION.md)
- [Deployment](docs/DEPLOYMENT.md)
- [Development](docs/DEVELOPMENT.md)

## License

MIT License - see LICENSE file for details
EOF

echo -e "  ${GREEN}✓${NC} README.md"

# LICENSE (MIT)
cat > LICENSE << EOF
MIT License

Copyright (c) $(date +%Y) $GITHUB_USER

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
EOF

echo -e "  ${GREEN}✓${NC} LICENSE"

# Basic docs
cat > docs/ARCHITECTURE.md << 'EOF'
# Gordon Watcher - Architecture

## Overview

Gordon Watcher is designed with production-grade patterns to handle high-volume file processing without overwhelming system resources.

## Components

### 1. File Watcher
- Uses `fsnotify` for efficient file system monitoring
- Recursive directory watching
- Automatic cleanup of empty directories

### 2. Worker Pool
- Fixed number of workers (prevents memory overflow)
- Buffered queue with backpressure
- Graceful shutdown

### 3. Rate Limiter
- Token bucket algorithm
- Protects downstream services
- Configurable rate per second

### 4. Circuit Breaker
- Protects against cascading failures
- Three states: Closed, Open, Half-Open
- Automatic recovery attempts

### 5. Storage Layer
- Redis for distributed deployments
- In-memory for single instance
- Idempotency guarantees

## Data Flow
```
File Detected → Worker Pool → Rate Limiter → Circuit Breaker → Queue/Storage
```

## Scaling

- Horizontal: Multiple watcher instances with Redis
- Vertical: Increase worker pool size
- Queue: RabbitMQ handles load distribution
EOF

echo -e "  ${GREEN}✓${NC} docs/ARCHITECTURE.md"

cat > docs/CONFIGURATION.md << 'EOF'
# Configuration Guide

## Configuration Files

Gordon Watcher supports multiple configuration sources (in order of priority):

1. Environment variables
2. Config file (YAML)
3. Default values

## Example Configuration

See `configs/config.example.yaml` for a complete example.

## Environment Variables

All configuration options can be set via environment variables with the prefix `GORDON_WATCHER_`.

Example:
```bash
GORDON_WATCHER_MAX_WORKERS=20
GORDON_WATCHER_LOG_LEVEL=debug
```

## Important Settings

### Worker Pool
- `max_workers`: Number of concurrent file processors (default: 10)
- `max_files_per_second`: Rate limit (default: 100)

### File Matching
- `file_patterns`: Files to process (e.g., ["*.xml", "*.zip"])
- `exclude_patterns`: Files to ignore (e.g., [".*", "*.tmp"])
EOF

echo -e "  ${GREEN}✓${NC} docs/CONFIGURATION.md"

# ═══════════════════════════════════════════════════════════
#  GIT INITIALIZATION
# ═══════════════════════════════════════════════════════════

echo ""
echo -e "${YELLOW}🔧 Initializing Git repository...${NC}"

if [ ! -d ".git" ]; then
    git init
    echo -e "  ${GREEN}✓${NC} Git initialized"
else
    echo -e "  ${YELLOW}⚠${NC} Git already initialized"
fi

# ═══════════════════════════════════════════════════════════
#  CREATE FILE TREE VISUALIZATION
# ═══════════════════════════════════════════════════════════

cat > STRUCTURE.md << 'EOF'
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
│   ├── config.dev.yaml
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
│   ├── banking/
│   └── generic/
├── Makefile                           # Build automation
├── go.mod                             # Go dependencies
└── README.md                          # Project overview
```
EOF

echo -e "  ${GREEN}✓${NC} STRUCTURE.md"

# ═══════════════════════════════════════════════════════════
#  SUMMARY
# ═══════════════════════════════════════════════════════════

echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║    GORDON WATCHER PROJECT CREATED! ✅             ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}Project Structure:${NC}"
echo -e "  📁 Directories: ${#directories[@]}"
echo -e "  📝 Files: ${#files[@]}"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo -e "  1. Review and complete the empty files"
echo -e "  2. Copy code from previous messages into respective files"
echo -e "  3. Run: ${BLUE}make deps${NC} (download dependencies)"
echo -e "  4. Run: ${BLUE}make build${NC} (build the project)"
echo -e "  5. Run: ${BLUE}make test${NC} (run tests)"
echo ""
echo -e "${YELLOW}Quick Commands:${NC}"
echo -e "  ${BLUE}make help${NC}         - Show all available commands"
echo -e "  ${BLUE}make dev${NC}          - Start development with hot reload"
echo -e "  ${BLUE}make test-stress${NC}  - Run stress tests"
echo ""
echo -e "${YELLOW}Documentation:${NC}"
echo -e "  📖 README.md           - Project overview"
echo -e "  📖 STRUCTURE.md        - Project structure"
echo -e "  📖 docs/               - Detailed documentation"
echo ""
echo -e "${GREEN}Happy coding! 🚀${NC}"