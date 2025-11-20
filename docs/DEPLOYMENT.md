# 🚀 Deployment Guide

Guia completo de deploy do Gordon Watcher em diferentes ambientes.

---

## 📋 Pré-requisitos

- Docker e Docker Compose instalados
- Acesso SSH ao servidor (para deploy com Ansible)
- Portas necessárias liberadas no firewall

---

## 🐳 Deploy com Docker Compose (Recomendado)

### 1. Preparar o Ambiente

```bash
# Clonar o repositório
git clone https://github.com/fabyo/gordon-watcher.git
cd gordon-watcher

# Criar arquivo .env com suas credenciais
cp .env.example .env
nano .env
```

**Configurar `.env`:**
```bash
RABBITMQ_USER=gordon
RABBITMQ_PASSWORD=sua_senha_segura_aqui
SAMBA_USER=gordon
SAMBA_PASSWORD=sua_senha_segura_aqui
DASHBOARD_PORT=8080
```

### 2. Criar Estrutura de Diretórios

```bash
sudo mkdir -p /opt/gordon-watcher/data/{incoming,processing,processed,failed}
sudo chown -R 1000:1000 /opt/gordon-watcher/data
sudo chmod -R 775 /opt/gordon-watcher/data
```

### 3. Iniciar os Serviços

```bash
# Build e start
docker compose up --build -d

# Verificar status
docker compose ps

# Verificar logs
docker compose logs -f watcher
```

### 4. Verificar Saúde

```bash
# Health check
curl http://localhost:8081/health

# Métricas
curl http://localhost:9100/metrics

# Dashboard
# Abra no navegador: http://localhost:8080
```

---

## 🤖 Deploy com Ansible

### 1. Configurar Inventário

Edite `ansible/inventory/hosts`:

```ini
[gordon_watcher]
servidor-producao ansible_host=192.168.1.100 ansible_user=ubuntu
```

### 2. Configurar Variáveis

Edite `ansible/group_vars/all.yml`:

```yaml
gordon_watcher_version: "v1.0.0"
gordon_user: gordon
gordon_group: gordon
gordon_watcher_install_dir: /opt/gordon-watcher
```

### 3. Executar Playbook

```bash
cd ansible

# Deploy completo
ansible-playbook -i inventory/hosts playbook.yml

# Apenas atualizar binário
ansible-playbook -i inventory/hosts playbook.yml --tags install

# Apenas configuração
ansible-playbook -i inventory/hosts playbook.yml --tags configure
```

### 4. Verificar Deploy

```bash
# SSH no servidor
ssh ubuntu@192.168.1.100

# Verificar serviço
sudo systemctl status gordon-watcher

# Ver logs
sudo journalctl -u gordon-watcher -f
```

---

## 🔄 Atualização de Versão

### Docker Compose

```bash
# Parar serviços
docker compose down

# Atualizar código
git pull

# Rebuild e restart
docker compose up --build -d
```

### Ansible

```bash
# Atualizar versão em group_vars/all.yml
gordon_watcher_version: "v1.1.0"

# Executar playbook
ansible-playbook -i inventory/hosts playbook.yml --tags install
```

---

## 🔐 Configuração de Segurança

### 1. Firewall

```bash
# Permitir apenas portas necessárias
sudo ufw allow 8080/tcp  # Dashboard
sudo ufw allow 4445/tcp  # Samba
sudo ufw enable
```

### 2. Senhas Fortes

Sempre use senhas fortes no `.env`:
```bash
# Gerar senha aleatória
openssl rand -base64 32
```

### 3. HTTPS (Opcional)

Para expor o dashboard via HTTPS, configure um reverse proxy (Nginx/Traefik) na frente do Caddy.

---

## 📊 Monitoramento em Produção

### Prometheus

Adicione o Gordon Watcher como target no `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'gordon-watcher'
    static_configs:
      - targets: ['servidor-producao:9100']
```

### Grafana

Importe dashboards prontos ou crie os seus usando as métricas:
- `gordon_watcher_files_detected_total`
- `gordon_watcher_files_sent_total`
- `gordon_watcher_files_rejected_total`

---

## 🔄 Backup e Restore

### Backup

```bash
# Backup dos dados
sudo tar -czf gordon-backup-$(date +%Y%m%d).tar.gz /opt/gordon-watcher/data

# Backup do RabbitMQ
docker exec gordon-rabbitmq rabbitmqctl export_definitions /tmp/rabbitmq-backup.json
docker cp gordon-rabbitmq:/tmp/rabbitmq-backup.json ./

# Backup do Redis
docker exec gordon-redis redis-cli SAVE
docker cp gordon-redis:/data/dump.rdb ./redis-backup.rdb
```

### Restore

```bash
# Restore dos dados
sudo tar -xzf gordon-backup-20250120.tar.gz -C /

# Restore do RabbitMQ
docker cp rabbitmq-backup.json gordon-rabbitmq:/tmp/
docker exec gordon-rabbitmq rabbitmqctl import_definitions /tmp/rabbitmq-backup.json

# Restore do Redis
docker cp redis-backup.rdb gordon-redis:/data/dump.rdb
docker compose restart redis
```

---

## 🆘 Rollback

Se algo der errado após atualização:

### Docker Compose

```bash
# Voltar para versão anterior
git checkout v1.0.0
docker compose up --build -d
```

### Ansible

O Ansible cria backup automático do binário. Para restaurar:

```bash
ssh servidor-producao
sudo cp /opt/gordon-watcher/bin/gordon-watcher.backup.TIMESTAMP /opt/gordon-watcher/bin/gordon-watcher
sudo systemctl restart gordon-watcher
```

---

## 📝 Checklist de Deploy

- [ ] Criar arquivo `.env` com credenciais seguras
- [ ] Criar estrutura de diretórios com permissões corretas
- [ ] Configurar firewall
- [ ] Iniciar serviços
- [ ] Verificar health check
- [ ] Verificar logs
- [ ] Testar processamento de arquivo
- [ ] Configurar monitoramento
- [ ] Configurar backup automático
- [ ] Documentar credenciais em local seguro
