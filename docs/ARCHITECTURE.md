# 🏗️ Arquitetura do Gordon Watcher

## Visão Geral

O Gordon Watcher foi projetado com padrões de nível de produção para lidar com processamento de arquivos de alto volume sem sobrecarregar os recursos do sistema.

## Componentes

### 1. File Watcher (Observador de Arquivos)
- Usa `fsnotify` para monitoramento eficiente do sistema de arquivos
- Monitoramento recursivo de diretórios
- Limpeza automática de diretórios vazios

### 2. Worker Pool (Pool de Trabalhadores)
- Número fixo de workers (previne overflow de memória)
- Fila com buffer e backpressure
- Desligamento gracioso

### 3. Rate Limiter (Limitador de Taxa)
- Algoritmo de token bucket
- Protege serviços downstream
- Taxa configurável por segundo

### 4. Circuit Breaker (Disjuntor)
- Protege contra falhas em cascata
- Três estados: Fechado, Aberto, Meio-Aberto
- Tentativas automáticas de recuperação

### 5. Camada de Storage
- Redis para deployments distribuídos
- In-memory para instância única
- Garantias de idempotência

## Fluxo de Dados
```
Arquivo Detectado → Worker Pool → Rate Limiter → Circuit Breaker → Queue/Storage
```

## Escalabilidade

- **Horizontal**: Múltiplas instâncias do watcher com Redis
- **Vertical**: Aumentar tamanho do worker pool
- **Queue**: RabbitMQ gerencia distribuição de carga
