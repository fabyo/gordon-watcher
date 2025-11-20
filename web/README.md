# 📊 Dashboard de Monitoramento

## 🎯 Acesso Rápido

Abra no navegador:
```
file:///home/fabyo/golang/gordon-watcher/web/dashboard.html
```

Ou via servidor HTTP simples:
```bash
cd /home/fabyo/golang/gordon-watcher/web
python3 -m http.server 8000
# Acesse: http://localhost:8000/dashboard.html
```

## 📈 Métricas Disponíveis

O dashboard mostra em tempo real:

### Arquivos
- **📥 Detectados**: Total de arquivos encontrados
- **✅ Enviados**: Processados com sucesso e enfileirados
- **🔄 Duplicados**: Já processados anteriormente (idempotência)
- **❌ Rejeitados**: Padrão ou tamanho inválido
- **🚫 Ignorados**: Arquivos ignorados por regras

### Sistema
- **⚡ Goroutines**: Threads ativas (detectar leaks)
- **👷 Workers Ativos**: Processando agora
- **📦 Fila**: Aguardando processamento

### Taxa de Processamento
- **Sucesso**: Percentual de arquivos processados com sucesso
- **Total**: Soma de todos os arquivos processados

## 🔄 Atualização

O dashboard atualiza automaticamente a cada **5 segundos**.

## 🔗 Links Rápidos

O dashboard inclui links para:
- 📊 Métricas Prometheus (http://localhost:9100/metrics)
- 🏥 Health Check (http://localhost:8081/health)
- 🐰 RabbitMQ Management (http://localhost:15672)
- 🔍 Jaeger Tracing (http://localhost:16686)

## 🎨 Personalização

Edite `web/dashboard.html` para:
- Mudar intervalo de atualização (linha 328)
- Adicionar novas métricas
- Customizar cores e layout
