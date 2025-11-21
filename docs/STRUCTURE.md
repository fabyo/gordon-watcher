# 📁 Estrutura do Projeto

```
gordon-watcher/
├── cmd/
│   └── watcher/
│       └── main.go                    # Ponto de entrada da aplicação
├── internal/
│   ├── config/                        # Gerenciamento de configuração
│   │   ├── config.go
│   │   ├── validator.go
│   │   └── defaults.go
│   ├── watcher/                       # Lógica principal do watcher
│   │   ├── watcher.go
│   │   ├── pool.go                    # Worker pool
│   │   └── doc.go
│   ├── queue/                         # Fila de mensagens
│   │   ├── queue.go
│   │   ├── message.go
│   │   ├── rabbitmq.go
│   │   └── noop.go
│   ├── storage/                       # Armazenamento de estado
│   │   ├── storage.go
│   │   ├── redis.go
│   │   └── memory.go
│   ├── health/                        # Health checks
│   │   └── health.go
│   ├── metrics/                       # Métricas Prometheus
│   │   ├── prometheus.go
│   │   └── server.go
│   └── telemetry/                     # OpenTelemetry
│       └── telemetry.go
├── configs/                           # Arquivos de configuração
│   ├── config.yaml
│   ├── config.example.yaml
│   ├── config.dev.yaml
│   └── config.test.yaml
├── ansible/                           # Automação de deploy
│   ├── playbook.yml
│   ├── deploy.yml
│   ├── rollback.yml
│   ├── inventory/
│   ├── roles/
│   │   └── gordon-watcher/
│   ├── group_vars/
│   └── scripts/
├── examples/                          # Usage examples
│   ├── banking/
│   └── generic/
├── docker/                            # Configurações Docker
│   ├── Dockerfile
│   └── Caddyfile
├── scripts/                           # Scripts auxiliares
│   ├── quick-test.sh
│   └── stress-test.sh
├── web/                               # Dashboard
│   ├── index.html
│   └── dashboard.png
├── docs/                              # Documentação
│   ├── ARCHITECTURE.md
│   ├── CONFIGURATION.md
│   ├── DEPLOYMENT.md
│   ├── DEVELOPMENT.md
│   ├── TROUBLESHOOTING.md
│   ├── GODOC.md
│   └── STRUCTURE.md                   # Este arquivo
├── .github/
│   └── workflows/
│       ├── build.yml
│       ├── test.yml
│       └── release.yml
├── docker-compose.yml                 # Stack completo
├── docker-compose.override.yml        # Overrides locais
├── Makefile                           # Automação de build
├── go.mod                             # Dependências Go
├── go.sum
├── LICENSE
└── README.md                          # Visão geral do projeto
```

## Diretórios Principais

### `/cmd`
Pontos de entrada da aplicação. Contém apenas o `main.go` que inicializa o watcher.

### `/internal`
Código interno da aplicação (não exportável). Contém toda a lógica de negócio.

### `/configs`
Arquivos de configuração YAML para diferentes ambientes.

### `/ansible`
Playbooks e roles para deploy automatizado com Ansible.

### `/docker`
Dockerfiles e configurações relacionadas ao Docker.

### `/scripts`
Scripts shell para testes, benchmarks e automação.

### `/web`
Dashboard HTML para monitoramento em tempo real.

### `/docs`
Documentação completa do projeto.

## Diretórios de Dados (Runtime)

Estes diretórios são criados em **runtime** (não estão no repositório):

```
/opt/gordon-watcher/data/
├── incoming/      # Arquivos detectados
├── processing/    # Arquivos sendo processados
├── processed/     # Arquivos processados com sucesso
├── failed/        # Arquivos que falharam
├── ignored/       # Arquivos ignorados
└── tmp/           # Temporários
```

**Nota:** A pasta `data/` **não existe** no repositório. Ela é criada pelo Ansible, Docker ou pelo script `setup.sh` durante a instalação.
