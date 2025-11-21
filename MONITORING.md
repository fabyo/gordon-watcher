# Gordon Watcher - Monitoramento

## Acessos via Caddy (Porta Única: 8080)

Todos os serviços de monitoramento estão acessíveis através do Caddy na porta configurada em `DASHBOARD_PORT` (padrão: 8080):

- **Dashboard Principal**: http://localhost:8080/
- **Grafana**: http://localhost:8080/grafana/
  - Usuário: `admin` (configurável via `GRAFANA_USER` no `.env`)
  - Senha: `admin` (configurável via `GRAFANA_PASSWORD` no `.env`)
- **Prometheus**: http://localhost:8080/prometheus/
  - Sem autenticação (acesso interno via Caddy)
- **Métricas (API)**: http://localhost:8080/api/metrics

## Acessos Diretos (Portas Individuais)

- **Jaeger UI**: http://localhost:16686
- **RabbitMQ Management**: http://localhost:15672
- **Watcher Health**: http://localhost:8081/health
- **Watcher Metrics**: http://localhost:9999/metrics

## Credenciais

Configure no arquivo `.env` (copie de `.env.example` se ainda não tiver):

### Grafana
```bash
GRAFANA_USER=admin
GRAFANA_PASSWORD=admin  # ⚠️ MUDE ISSO EM PRODUÇÃO!
```

### RabbitMQ
```bash
RABBITMQ_USER=gordon
RABBITMQ_PASSWORD=secret_change_me  # ⚠️ MUDE ISSO EM PRODUÇÃO!
```

### Samba
```bash
SAMBA_USER=gordon
SAMBA_PASSWORD=secret_change_me  # ⚠️ MUDE ISSO EM PRODUÇÃO!
```

> **Nota**: O Prometheus não tem autenticação por padrão. Ele está acessível apenas internamente via Caddy.

## Alterando a Porta do Dashboard

Para mudar a porta do Caddy, edite o `.env`:

```bash
DASHBOARD_PORT=9000
```

Depois reinicie:

```bash
docker compose restart dashboard grafana
```

O Grafana será automaticamente reconfigurado para a nova porta.
