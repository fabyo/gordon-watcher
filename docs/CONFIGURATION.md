# ⚙️ Guia de Configuração

## Arquivos de Configuração

O Gordon Watcher suporta múltiplas fontes de configuração (em ordem de prioridade):

1. Variáveis de ambiente
2. Arquivo de configuração (YAML)
3. Valores padrão

## Exemplo de Configuração

Veja `configs/config.example.yaml` para um exemplo completo.

## Variáveis de Ambiente

Todas as opções de configuração podem ser definidas via variáveis de ambiente com o prefixo `GORDON_WATCHER_`.

Exemplo:
```bash
GORDON_WATCHER_MAX_WORKERS=20
GORDON_WATCHER_LOG_LEVEL=debug
```

## Configurações Importantes

### Worker Pool
- `max_workers`: Número de processadores de arquivo concorrentes (padrão: 10)
- `max_files_per_second`: Limite de taxa (padrão: 100)

### Correspondência de Arquivos
- `file_patterns`: Arquivos a processar (ex: ["*.xml", "*.zip"])
- `exclude_patterns`: Arquivos a ignorar (ex: [".*", "*.tmp"])
