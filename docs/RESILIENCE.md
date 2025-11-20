# Funcionalidades de Resiliência e Confiabilidade

O Gordon Watcher foi projetado com foco em alta disponibilidade e integridade de dados. Abaixo estão detalhadas as principais funcionalidades que garantem a robustez do sistema.

## 1. Idempotência (Idempotency)
Garante que o mesmo arquivo não seja processado mais de uma vez, mesmo que eventos duplicados sejam recebidos.
- **Como funciona:** O sistema calcula um hash SHA-256 do conteúdo do arquivo + nome do arquivo. Antes de processar, verifica no Redis se esse hash já foi marcado como processado.
- **Benefício:** Evita duplicação de dados e processamento redundante.

## 2. Bloqueio Distribuído (Distributed Locking)
Impede que múltiplos workers (em diferentes containers ou threads) processem o mesmo arquivo simultaneamente.
- **Como funciona:** Utiliza o Redis para adquirir um "lock" temporário baseado no hash do arquivo. Se um worker tentar pegar um arquivo bloqueado, ele ignora e passa para o próximo.
- **Benefício:** Garante consistência e evita condições de corrida.

## 3. Lógica de Retentativa (Retry Logic)
Trata falhas transitórias de forma automática.
- **Como funciona:** Se o processamento falhar (ex: erro de rede, timeout), o arquivo é re-enfileirado para uma nova tentativa. O número de tentativas é configurável.
- **Benefício:** Aumenta a taxa de sucesso sem intervenção manual.

## 4. Circuit Breaker
Protege o sistema contra falhas em cascata quando serviços dependentes (RabbitMQ, Redis, Samba) estão indisponíveis.
- **Como funciona:** Monitora a taxa de erros. Se ultrapassar um limite, o sistema para temporariamente de enviar requisições para o serviço afetado, permitindo que ele se recupere.
- **Benefício:** Evita sobrecarga em sistemas que já estão falhando.

## 5. Dead Letter Queue (DLQ) - Fila de Falhas
Armazena arquivos que não puderam ser processados após todas as tentativas.
- **Como funciona:** Arquivos que excedem o limite de retentativas são movidos para a pasta `failed`.
- **Benefício:** Permite isolar problemas e facilita a análise manual sem travar o fluxo principal.

## 6. Graceful Shutdown
Garante que o sistema desligue de forma segura, sem perder dados em processamento.
- **Como funciona:** Ao receber um sinal de parada (SIGTERM/SIGINT), o watcher para de aceitar novos arquivos, espera os workers terminarem as tarefas atuais e fecha as conexões corretamente.
- **Benefício:** Previne corrupção de dados durante deploys ou reinícios.

## 7. Entrega "At-Least-Once" (Pelo menos uma vez)
Garante que nenhum arquivo seja perdido.
- **Como funciona:** O arquivo só é considerado "processado" após a confirmação de sucesso. Se o worker cair no meio do processo, o arquivo permanece na fila ou é reprocessado na próxima inicialização (via scan inicial).
- **Benefício:** Confiabilidade total no processamento de dados críticos.

## 8. Extração Automática de ZIP
Processa arquivos compactados de forma transparente.
- **Como funciona:** Detecta arquivos `.zip`, extrai o conteúdo para a pasta de processamento e remove o arquivo original. Protegido contra vulnerabilidades "ZipSlip".
- **Benefício:** Facilita a integração com sistemas que enviam lotes de arquivos compactados.

## 9. Limpeza Automática (Cleanup Scheduler)
Gerencia o ciclo de vida dos arquivos processados e falhos.
- **Como funciona:** Um agendador (cron) roda periodicamente para limpar arquivos antigos nas pastas `processed`, `failed`, `ignored` e `tmp`, baseado em políticas de retenção configuráveis.
- **Benefício:** Mantém o disco limpo e o sistema organizado automaticamente.
