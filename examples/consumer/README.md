# Gordon Watcher Consumer Example

Este é um exemplo de consumer (worker) que processa mensagens da fila RabbitMQ populada pelo Gordon Watcher.

## 📋 O Que Faz

O consumer:
- Conecta ao RabbitMQ
- Consome mensagens da fila `xml` (ou outra configurada)
- Processa cada arquivo de forma assíncrona
- Faz ACK/NACK apropriado
- Suporta graceful shutdown

## 🚀 Como Usar

### Localmente

```bash
# Instalar dependências
cd examples/consumer
go mod init consumer
go get github.com/rabbitmq/amqp091-go

# Rodar
export RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
export QUEUE_NAME="xml"
go run main.go
```

### Com Docker

```bash
# Build
docker build -t gordon-consumer -f examples/consumer/Dockerfile .

# Run
docker run --rm \
  -e RABBITMQ_URL="amqp://guest:guest@rabbitmq:5672/" \
  -e QUEUE_NAME="xml" \
  --network gordon-watcher_default \
  gordon-consumer
```

### Com Docker Compose

Adicione ao seu `docker-compose.yml`:

```yaml
  consumer:
    build:
      context: .
      dockerfile: examples/consumer/Dockerfile
    environment:
      RABBITMQ_URL: amqp://guest:guest@rabbitmq:5672/
      QUEUE_NAME: xml
    depends_on:
      - rabbitmq
    restart: unless-stopped
```

## 🔧 Configuração

Variáveis de ambiente:

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | URL de conexão RabbitMQ |
| `QUEUE_NAME` | `xml` | Nome da fila para consumir |

## 📦 Estrutura da Mensagem

Cada mensagem recebida tem o seguinte formato JSON:

```json
{
  "path": "/data/processing/file.xml",
  "hash": "abc123...",
  "size": 1024,
  "timestamp": "2024-01-01T12:00:00Z",
  "queue": "xml"
}
```

## 💡 Implementando Sua Lógica

Edite a função `processMessage()` em `main.go`:

```go
func (c *Consumer) processMessage(ctx context.Context, delivery amqp.Delivery) error {
    var msg Message
    if err := json.Unmarshal(delivery.Body, &msg); err != nil {
        return fmt.Errorf("failed to unmarshal message: %w", err)
    }

    // SUA LÓGICA AQUI
    // Exemplos:
    // - Ler arquivo do path
    // - Parsear XML/JSON
    // - Salvar em banco de dados
    // - Chamar API externa
    // - Gerar relatórios
    // - Etc.

    return nil
}
```

## 🔄 Fluxo de Processamento

1. **Recebe mensagem** da fila
2. **Processa** o arquivo (sua lógica)
3. **ACK** se sucesso → mensagem removida da fila
4. **NACK** se erro → mensagem vai para DLQ (se configurado)

## 🏥 Health & Monitoring

Para produção, adicione:
- Health check endpoint
- Métricas Prometheus
- Logging estruturado
- Tracing distribuído

## 📚 Próximos Passos

1. Implementar sua lógica de negócio
2. Adicionar testes
3. Configurar retry policy
4. Adicionar métricas
5. Deploy em produção
