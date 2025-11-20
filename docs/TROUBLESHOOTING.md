# 🔧 Troubleshooting

Guia de solução de problemas comuns do Gordon Watcher.

---

## 🚫 Watcher não está processando arquivos

### Sintoma
Arquivos ficam parados na pasta `incoming` e não são movidos.

### Diagnóstico

1. **Verificar logs em tempo real:**
   ```bash
   docker compose logs -f watcher
   ```

2. **Verificar se o container vê os arquivos:**
   ```bash
   docker exec -it gordon-watcher ls -l /opt/gordon-watcher/data/incoming
   ```

3. **Testar permissões de escrita:**
   ```bash
   docker exec -it gordon-watcher touch /opt/gordon-watcher/data/incoming/teste.xml
   ```

### Soluções Comuns

#### Permissões incorretas
```bash
sudo chown -R 1000:1000 /opt/gordon-watcher/data
sudo chmod -R 775 /opt/gordon-watcher/data
```

#### Sistema de arquivos não suporta inotify
Alguns sistemas (VirtualBox shares, NFS) não enviam eventos de notificação. Teste com:
```bash
sudo apt-get install inotify-tools -y
inotifywait -m /opt/gordon-watcher/data/incoming
```

Se não aparecer nada ao copiar arquivos, o sistema de arquivos não é compatível.

---

## 🔒 Erros de "lock already held"

### Sintoma
Logs mostram: `Failed to acquire lock (another worker processing?)`

### Causa
Múltiplos arquivos com conteúdo idêntico sendo processados simultaneamente.

### Solução
Isso é comportamento esperado para arquivos duplicados. O watcher usa hash de conteúdo + nome do arquivo para evitar reprocessamento.

Se você está vendo isso com arquivos diferentes, verifique se não há múltiplas instâncias do watcher rodando:
```bash
pkill -f gordon-watcher
docker compose restart watcher
```

---

## 🌐 Dashboard não carrega / Erro 404

### Sintoma
Ao acessar `http://localhost:8080`, recebe erro 404 ou "Cannot connect".

### Soluções

1. **Verificar se o container está rodando:**
   ```bash
   docker ps | grep gordon-dashboard
   ```

2. **Verificar logs do Caddy:**
   ```bash
   docker compose logs dashboard
   ```

3. **Porta ocupada:**
   ```bash
   # Verificar se a porta 8080 está livre
   sudo netstat -tulpn | grep 8080
   
   # Usar porta alternativa
   DASHBOARD_PORT=9090 docker compose up -d
   ```

---

## 🐰 RabbitMQ não conecta

### Sintoma
Logs mostram: `Failed to connect to RabbitMQ` ou `connection refused`.

### Soluções

1. **Verificar se o RabbitMQ está rodando:**
   ```bash
   docker ps | grep rabbitmq
   docker compose logs rabbitmq
   ```

2. **Verificar credenciais:**
   Certifique-se de que o arquivo `.env` existe e contém:
   ```bash
   RABBITMQ_USER=gordon
   RABBITMQ_PASSWORD=sua_senha_aqui
   ```

3. **Recriar containers:**
   ```bash
   docker compose down -v
   docker compose up -d
   ```

---

## 🔴 Redis não conecta

### Sintoma
Logs mostram: `Failed to connect to Redis`.

### Soluções

1. **Verificar se o Redis está rodando:**
   ```bash
   docker ps | grep redis
   docker compose logs redis
   ```

2. **Testar conexão manualmente:**
   ```bash
   docker exec -it gordon-redis redis-cli ping
   # Deve retornar: PONG
   ```

---

## 🛑 Parar o Watcher Forçadamente

Se o watcher travar ou você precisar matá-lo rapidamente:

```bash
# Se rodando via Docker
docker compose stop watcher

# Se rodando diretamente (sem Docker)
pkill -f gordon-watcher
```

---

## 🔍 Verificar Portas Ocupadas

Para verificar se as portas necessárias estão livres:

```bash
sudo netstat -tulpn | grep -E '8080|8081|9100|5672|15672|6379|16686|14268|139|4445'
```

**Portas usadas:**
- `8080` - Dashboard (Caddy)
- `8081` - Health Check
- `9100` - Métricas (Prometheus)
- `5672` - RabbitMQ AMQP
- `15672` - RabbitMQ Management UI
- `6379` - Redis
- `16686` - Jaeger UI
- `14268` - Jaeger Collector
- `139` - Samba
- `4445` - Samba (mapeado de 445)

---

## 📧 Muitos emails do GitHub Actions

Se você está recebendo muitos emails de build:

O workflow já está configurado para ignorar mudanças em arquivos `.md`. Se ainda receber emails, verifique as configurações de notificação do GitHub:

1. Acesse: `https://github.com/settings/notifications`
2. Em "Actions", ajuste para "Only notify on failures"

---

## 🆘 Precisa de mais ajuda?

1. Verifique os logs completos: `docker compose logs`
2. Abra uma issue no GitHub: https://github.com/fabyo/gordon-watcher/issues
3. Inclua sempre:
   - Versão do Gordon Watcher (`docker exec gordon-watcher /app/gordon-watcher --version`)
   - Sistema operacional
   - Logs relevantes
