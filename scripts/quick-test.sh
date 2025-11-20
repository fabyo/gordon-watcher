#!/bin/bash
# Script de teste rápido do Gordon Watcher

echo "🧪 Teste Rápido do Gordon Watcher"
echo "================================="
echo ""

# 1. Criar pastas de teste no home
TEST_DIR="$HOME/gordon-test"
echo "📁 Criando pastas em $TEST_DIR..."
mkdir -p $TEST_DIR/{incoming,processing,processed,failed,ignored,tmp}

# Backup config original
if [ -f configs/config.yaml ]; then
    cp configs/config.yaml configs/config.yaml.bak
fi

# 2. Criar config de teste (sobrescrevendo o padrão pois o watcher lê configs/config.yaml)
echo "⚙️  Configurando..."
cat > configs/config.yaml <<EOF
app:
  name: gordon-watcher
  version: dev
  environment: test

watcher:
  paths:
    - $TEST_DIR/incoming
  file_patterns:
    - "*.xml"
    - "*.txt"
  min_file_size: 1
  max_file_size: 104857600
  stable_attempts: 2
  stable_delay: 100000000 # 100ms in nanoseconds
  cleanup_interval: 300000000000 # 5m in nanoseconds
  max_workers: 3
  max_files_per_second: 10
  worker_queue_size: 5
  working_dir: $TEST_DIR
  sub_directories:
    processing: processing
    processed: processed
    failed: failed
    ignored: ignored
    tmp: tmp

queue:
  enabled: false

redis:
  enabled: false

metrics:
  addr: :9100

health:
  addr: :8081

telemetry:
  enabled: false

logger:
  level: info
  format: text
  output: stdout
EOF

# 3. Rodar o watcher
echo "🚀 Iniciando Gordon Watcher..."
echo ""
./bin/gordon-watcher &
WATCHER_PID=$!

sleep 2

# 4. Criar arquivo de teste
echo "📝 Criando arquivo de teste..."
echo '<?xml version="1.0"?><nota>Teste Gordon Watcher</nota>' > $TEST_DIR/incoming/teste-$(date +%s).xml

sleep 1

# 5. Verificar
echo ""
echo "✅ Verificando resultados:"
echo "  - Incoming: $(ls -1 $TEST_DIR/incoming 2>/dev/null | wc -l) arquivos"
echo "  - Processing: $(ls -1 $TEST_DIR/processing 2>/dev/null | wc -l) arquivos"
echo ""
echo "📊 Métricas disponíveis em: http://localhost:9100/metrics"
echo "🏥 Health check em: http://localhost:8081/health"
echo ""
echo "Para parar o watcher: kill $WATCHER_PID"
echo "PID do watcher: $WATCHER_PID"

# Restaurar config original ao sair (opcional, mas bom para limpeza)
# trap "mv configs/config.yaml.bak configs/config.yaml" EXIT
