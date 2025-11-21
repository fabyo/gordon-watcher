# 👨‍💻 Guia de Desenvolvimento

Guia para desenvolvedores que querem contribuir com o Gordon Watcher.

---

## 🛠️ Configuração do Ambiente

### Pré-requisitos

- Go 1.21 ou superior
- Docker e Docker Compose
- Make
- Git

### Instalação

```bash
# Clonar o repositório
git clone https://github.com/fabyo/gordon-watcher.git
cd gordon-watcher

# Instalar dependências
go mod download

# Build
make build

# Rodar testes
make test
```

---

## 📁 Estrutura do Projeto

```
gordon-watcher/
├── cmd/
│   └── watcher/          # Ponto de entrada da aplicação
│       └── main.go
├── internal/
│   ├── config/           # Configuração
│   ├── health/           # Health check HTTP
│   ├── metrics/          # Métricas Prometheus
│   ├── queue/            # Abstração de filas (RabbitMQ/NoOp)
│   ├── storage/          # Abstração de storage (Redis/Memory)
│   ├── telemetry/        # OpenTelemetry/OTLP (Jaeger compatible)
│   └── watcher/          # Lógica principal do watcher
│       ├── watcher.go    # Core do watcher
│       └── pool.go       # Worker pool
├── configs/              # Arquivos de configuração
├── docker/               # Dockerfiles e configs
├── ansible/              # Playbooks de deploy
├── scripts/              # Scripts utilitários
└── web/                  # Dashboard HTML
```

---

## 🔧 Comandos Make

```bash
# Build
make build              # Compilar binário
make build-docker       # Build da imagem Docker

# Testes
make test               # Rodar testes
make test-coverage      # Testes com cobertura
make test-integration   # Testes de integração

# Qualidade de Código
make lint               # Linter (golangci-lint)
make fmt                # Formatar código

# Desenvolvimento
make run                # Rodar localmente
make dev                # Rodar com hot-reload
make clean              # Limpar builds

# Docker
make docker-up          # Subir stack completo
make docker-down        # Parar stack
make docker-logs        # Ver logs

# Utilitários
make discover-ip        # Descobrir IP da máquina
```

---

## 🧪 Testes

### Testes Unitários

```bash
# Rodar todos os testes
make test

# Rodar testes de um pacote específico
go test ./internal/watcher -v

# Rodar com cobertura
make test-coverage
```

### Testes de Integração

```bash
# Rodar testes de integração
make test-integration

# Ou manualmente
go test -tags=integration ./...
```

### Teste Rápido

Use o script `quick-test.sh` para testar o fluxo completo:

```bash
./scripts/quick-test.sh
```

Ele vai:
1. Criar estrutura de diretórios temporária
2. Gerar config de teste
3. Iniciar o watcher
4. Criar arquivo XML de teste
5. Verificar se foi processado

---

## 🏗️ Arquitetura

### Fluxo de Processamento

```
1. fsnotify detecta arquivo novo
2. Watcher aguarda estabilização (arquivo parou de crescer)
3. Calcula hash (SHA256 de conteúdo + nome)
4. Tenta adquirir lock distribuído (Redis/Memory)
5. Move para /processing
6. Envia para fila (RabbitMQ/NoOp)
7. Move para /processed
8. Libera lock
```

### Componentes Principais

#### Watcher (`internal/watcher/watcher.go`)
- Monitora diretórios com `fsnotify`
- Gerencia ciclo de vida dos arquivos
- Coordena workers

#### Worker Pool (`internal/watcher/pool.go`)
- Pool de goroutines para processar arquivos
- Controla concorrência
- Gerencia fila interna

#### Storage (`internal/storage/`)
- Interface para locks distribuídos
- Implementações: Redis e In-Memory

#### Queue (`internal/queue/`)
- Interface para filas de mensagens
- Implementações: RabbitMQ e NoOp

---

## 🔌 Adicionando Novas Features

### 1. Criar Issue

Antes de começar, crie uma issue no GitHub descrevendo a feature.

### 2. Criar Branch

```bash
git checkout -b feature/nome-da-feature
```

### 3. Implementar

Siga os padrões do projeto:
- Use interfaces para abstrações
- Adicione testes
- Documente funções públicas
- Use structured logging

### 4. Testar

```bash
make test
make lint
```

### 5. Commit

Use commits semânticos:
```bash
git commit -m "feat: adiciona suporte para arquivos PDF"
git commit -m "fix: corrige race condition no worker pool"
git commit -m "docs: atualiza README com exemplos"
```

### 6. Pull Request

Abra um PR com:
- Descrição clara da mudança
- Link para a issue
- Screenshots (se aplicável)
- Checklist de testes

---

## 📝 Convenções de Código

### Naming

- Pacotes: lowercase, sem underscores
- Funções públicas: PascalCase
- Funções privadas: camelCase
- Constantes: PascalCase ou UPPER_CASE

### Logging

Use structured logging:

```go
logger.Info("Processing file",
    "path", filePath,
    "size", fileSize,
    "hash", hash)
```

### Errors

Sempre adicione contexto aos erros:

```go
return fmt.Errorf("failed to process file %s: %w", path, err)
```

---

## 🐛 Debugging

### Logs Detalhados

```bash
# Rodar com log level debug
LOG_LEVEL=debug ./bin/gordon-watcher
```

### Delve (Debugger)

```bash
# Instalar
go install github.com/go-delve/delve/cmd/dlv@latest

# Debugar
dlv debug ./cmd/watcher
```

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./internal/watcher
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./internal/watcher
go tool pprof mem.prof
```

---

## 📦 Release

### Criar Nova Versão

```bash
# Atualizar versão no código
# Criar tag
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0
```

O GitHub Actions vai automaticamente:
1. Rodar testes
2. Fazer build para Linux/AMD64 e ARM64
3. Criar release no GitHub
4. Anexar binários

---

## 🤝 Contribuindo

1. Fork o projeto
2. Crie sua feature branch
3. Commit suas mudanças
4. Push para a branch
5. Abra um Pull Request

Leia o [CONTRIBUTING.md](../CONTRIBUTING.md) para mais detalhes.

---

## 📚 Recursos Úteis

- [Go Documentation](https://go.dev/doc/)
- [fsnotify](https://github.com/fsnotify/fsnotify)
- [RabbitMQ Go Client](https://github.com/rabbitmq/amqp091-go)
- [Redis Go Client](https://github.com/redis/go-redis)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [Prometheus Client](https://github.com/prometheus/client_golang)
