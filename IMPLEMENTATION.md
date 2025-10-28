# Detalhes da Implementação - Fechamento Automático de Leilões

## 📋 Visão Geral

Este documento detalha a implementação da funcionalidade de fechamento automático de leilões usando Go routines e concorrência segura.

## 🎯 Requisitos Implementados

✅ Função que calcula o tempo do leilão baseado em variáveis de ambiente  
✅ Go routine que valida leilões expirados e realiza o update  
✅ Testes automatizados para validar o fechamento  
✅ Solução thread-safe com controle de concorrência  
✅ Docker e docker-compose para ambiente de desenvolvimento  

## 🔧 Implementação Técnica

### 1. Cálculo de Duração do Leilão

**Arquivo:** `internal/infra/database/auction/create_auction.go`

```go
func calculateAuctionDuration() time.Duration {
    durationStr := os.Getenv("AUCTION_DURATION")
    if durationStr == "" {
        return 5 * time.Minute // Padrão: 5 minutos
    }

    durationSeconds, err := strconv.Atoi(durationStr)
    if err != nil {
        log.Printf("Invalid AUCTION_DURATION value, using default 5 minutes")
        return 5 * time.Minute
    }

    return time.Duration(durationSeconds) * time.Second
}
```

**Características:**
- Lê a variável de ambiente `AUCTION_DURATION`
- Valor padrão de 5 minutos se não configurado
- Tratamento de erro para valores inválidos
- Retorna `time.Duration` para uso consistente

### 2. Inicialização da Go Routine

**Arquivo:** `internal/infra/database/auction/create_auction.go`

```go
func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
    repo := &AuctionRepository{
        Collection: database.Collection("auctions"),
    }

    // Inicia a goroutine para verificar leilões expirados
    go repo.startAuctionExpirationChecker()

    return repo
}
```

**Características:**
- Go routine iniciada automaticamente na criação do repository
- Executa em background durante toda a vida da aplicação
- Não bloqueia a inicialização da aplicação

### 3. Verificador de Expiração

**Arquivo:** `internal/infra/database/auction/create_auction.go`

```go
func (ar *AuctionRepository) startAuctionExpirationChecker() {
    checkInterval := getCheckInterval()
    ticker := time.NewTicker(checkInterval)
    defer ticker.Stop()

    log.Printf("Auction expiration checker started with interval: %v", checkInterval)

    for range ticker.C {
        ctx := context.Background()
        if err := ar.closeExpiredAuctions(ctx); err != nil {
            log.Printf("Error closing expired auctions: %v", err)
        }
    }
}
```

**Características:**
- Usa `time.Ticker` para execução periódica
- Intervalo configurável via `AUCTION_CHECK_INTERVAL`
- Tratamento de erro sem interromper a goroutine
- Logging para monitoramento

### 4. Fechamento de Leilões Expirados

**Arquivo:** `internal/infra/database/auction/create_auction.go`

```go
func (ar *AuctionRepository) closeExpiredAuctions(ctx context.Context) error {
    ar.mu.Lock()
    defer ar.mu.Unlock()

    now := time.Now().Unix()

    // Busca leilões ativos que já expiraram
    filter := bson.M{
        "status":     entity.Active,
        "expires_at": bson.M{"$lte": now},
    }

    cursor, err := ar.Collection.Find(ctx, filter)
    if err != nil {
        return err
    }
    defer cursor.Close(ctx)

    var expiredAuctions []entity.AuctionEntityMongo
    if err := cursor.All(ctx, &expiredAuctions); err != nil {
        return err
    }

    // Fecha cada leilão expirado
    for _, auction := range expiredAuctions {
        update := bson.M{
            "$set": bson.M{
                "status": entity.Completed,
            },
        }

        _, err := ar.Collection.UpdateOne(ctx, bson.M{"_id": auction.Id}, update)
        if err != nil {
            log.Printf("Error updating auction %s status: %v", auction.Id, err)
            continue
        }

        log.Printf("Auction %s closed automatically (expired at: %s)", 
            auction.Id, 
            time.Unix(auction.ExpiresAt, 0).Format(time.RFC3339))
    }

    if len(expiredAuctions) > 0 {
        log.Printf("Closed %d expired auction(s)", len(expiredAuctions))
    }

    return nil
}
```

**Características:**
- Query eficiente no MongoDB (índice em `status` e `expires_at`)
- Atualização em lote de leilões expirados
- Logging detalhado de cada fechamento
- Continua processando mesmo se um update falhar

### 5. Controle de Concorrência

**Arquivo:** `internal/infra/database/auction/create_auction.go`

```go
type AuctionRepository struct {
    Collection *mongo.Collection
    mu         sync.RWMutex
}

// Operações de leitura
func (ar *AuctionRepository) FindAuctionById(ctx context.Context, id string) (*entity.Auction, error) {
    ar.mu.RLock()
    defer ar.mu.RUnlock()
    // ... código de busca
}

// Operações de escrita
func (ar *AuctionRepository) CreateAuction(ctx context.Context, auction *entity.Auction) error {
    ar.mu.Lock()
    defer ar.mu.Unlock()
    // ... código de criação
}
```

**Características:**
- `sync.RWMutex` para controle de acesso
- `RLock/RUnlock` para operações de leitura (múltiplas simultâneas)
- `Lock/Unlock` para operações de escrita (exclusivas)
- Previne race conditions em ambiente concorrente

### 6. Validação de Leilão Ativo

**Arquivo:** `internal/infra/database/bid/create_bid.go`

```go
func (br *BidRepository) CreateBid(ctx context.Context, bid *entity.Bid) error {
    // Valida se o leilão existe e está ativo
    auction, err := br.AuctionRepository.FindAuctionById(ctx, bid.AuctionId)
    if err != nil {
        return err
    }

    if auction == nil {
        return errors.New("auction not found")
    }

    // Verifica se o leilão está ativo
    if auction.Status != entity.Active {
        return errors.New("auction is not active")
    }

    // Verifica se o leilão expirou
    if auction.IsExpired() {
        return errors.New("auction has expired")
    }

    // ... continua com criação do lance
}
```

**Características:**
- Validação em três níveis: existência, status e expiração
- Previne lances em leilões fechados
- Mensagens de erro claras

## 🧪 Testes Implementados

### Teste Principal: TestAuctionAutomaticClosure

**Arquivo:** `internal/infra/database/auction/create_auction_test.go`

```go
func TestAuctionAutomaticClosure(t *testing.T) {
    database, cleanup := setupTestDB(t)
    defer cleanup()

    // Define duração curta para teste (3 segundos)
    os.Setenv("AUCTION_DURATION", "3")
    os.Setenv("AUCTION_CHECK_INTERVAL", "1")
    defer os.Unsetenv("AUCTION_DURATION")
    defer os.Unsetenv("AUCTION_CHECK_INTERVAL")

    repo := NewAuctionRepository(database)
    ctx := context.Background()

    // Cria um leilão
    auction, err := entity.CreateAuction(
        "Test Product",
        "Test Category",
        "This is a test product for automatic closure",
        entity.New,
        0,
    )
    assert.NoError(t, err)
    err = repo.CreateAuction(ctx, auction)
    assert.NoError(t, err)

    // Verifica que o leilão está ativo
    foundAuction, err := repo.FindAuctionById(ctx, auction.Id)
    assert.NoError(t, err)
    assert.Equal(t, entity.Active, foundAuction.Status)

    // Aguarda o leilão expirar + tempo para a goroutine processar
    time.Sleep(5 * time.Second)

    // Verifica que o leilão foi fechado automaticamente
    closedAuction, err := repo.FindAuctionById(ctx, auction.Id)
    assert.NoError(t, err)
    assert.NotNil(t, closedAuction)
    assert.Equal(t, entity.Completed, closedAuction.Status)
}
```

**O que o teste valida:**
1. ✅ Leilão é criado com status Active
2. ✅ Tempo de expiração é calculado corretamente
3. ✅ Go routine detecta o leilão expirado
4. ✅ Status é atualizado para Completed automaticamente
5. ✅ Processo ocorre sem intervenção manual

### Outros Testes

**TestConcurrentAuctionCreation**
- Valida criação simultânea de múltiplos leilões
- Garante thread-safety do mutex

**TestCloseExpiredAuctionsDirectly**
- Testa o método de fechamento diretamente
- Valida query e update no MongoDB

**TestCalculateAuctionDuration**
- Testa cálculo de duração com diferentes valores
- Valida tratamento de valores inválidos

## 📊 Fluxo de Dados

```
1. Aplicação Inicia
   ↓
2. AuctionRepository é criado
   ↓
3. Go routine startAuctionExpirationChecker() inicia
   ↓
4. Usuário cria leilão via API
   ↓
5. CreateAuction calcula expires_at = now + AUCTION_DURATION
   ↓
6. Leilão salvo no MongoDB com status=Active
   ↓
7. [A cada AUCTION_CHECK_INTERVAL segundos]
   ↓
8. closeExpiredAuctions() executa
   ↓
9. Query: status=Active AND expires_at <= now
   ↓
10. Para cada leilão expirado:
    - Update status = Completed
    - Log do fechamento
   ↓
11. Novos lances são rejeitados (status != Active)
```

## 🔒 Considerações de Segurança e Performance

### Concorrência
- ✅ Mutex protege operações críticas
- ✅ RWMutex permite múltiplas leituras simultâneas
- ✅ Locks são sempre liberados (defer)

### Performance
- ✅ Query indexada no MongoDB (status + expires_at)
- ✅ Intervalo de verificação configurável
- ✅ Processamento em lote de leilões expirados

### Confiabilidade
- ✅ Goroutine não para em caso de erro
- ✅ Logging completo para debugging
- ✅ Tratamento de edge cases

### Escalabilidade
- ⚠️ Solução atual: single-instance
- 💡 Para produção: considerar sistema de filas (RabbitMQ, Kafka)
- 💡 Alternativa: usar MongoDB Change Streams
- 💡 Alternativa: usar scheduled jobs (cron)

## 🎓 Conceitos de Go Utilizados

1. **Goroutines**: Concorrência leve
2. **Channels**: Comunicação via ticker
3. **Mutex**: Sincronização de acesso
4. **Context**: Controle de timeout e cancelamento
5. **Defer**: Garantia de liberação de recursos
6. **Interfaces**: Desacoplamento de dependências

## 📈 Métricas e Monitoramento

### Logs Importantes

```
# Inicialização
Auction expiration checker started with interval: 10s

# Fechamento
Auction abc-123 closed automatically (expired at: 2024-01-15T10:05:00Z)
Closed 1 expired auction(s)

# Erros
Error closing expired auctions: connection timeout
Error updating auction xyz-456 status: document not found
```

### Recomendações para Produção

1. Adicionar métricas Prometheus
2. Implementar health checks
3. Configurar alertas para falhas
4. Adicionar distributed tracing
5. Implementar circuit breaker

## 🚀 Melhorias Futuras

1. **Notificações**: Enviar email/webhook quando leilão fechar
2. **Histórico**: Manter log de mudanças de status
3. **Extensão de tempo**: Permitir estender leilão
4. **Leilão recorrente**: Reabrir automaticamente
5. **Priorização**: Processar leilões de alto valor primeiro

## 📚 Referências

- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [MongoDB Go Driver](https://www.mongodb.com/docs/drivers/go/current/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Testing](https://go.dev/doc/tutorial/add-a-test)
