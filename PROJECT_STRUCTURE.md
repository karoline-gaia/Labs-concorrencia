# 📁 Estrutura do Projeto

```
auction-goexpert/
│
├── cmd/
│   └── auction/
│       └── main.go                          # Ponto de entrada da aplicação
│
├── configuration/
│   └── database/
│       └── mongodb/
│           └── connection.go                # Configuração e conexão MongoDB
│
├── internal/
│   ├── entity/                              # Entidades de domínio
│   │   ├── auction_entity.go                # Entidade Auction + interfaces
│   │   ├── bid_entity.go                    # Entidade Bid + interfaces
│   │   └── user_entity.go                   # Entidade User + interfaces
│   │
│   ├── usecase/                             # Casos de uso (regras de negócio)
│   │   ├── auction_usecase/
│   │   │   ├── create_auction_usecase.go    # UC: Criar leilão
│   │   │   └── find_auction_usecase.go      # UC: Buscar leilões
│   │   └── bid_usecase/
│   │       ├── create_bid_usecase.go        # UC: Criar lance
│   │       └── find_bid_usecase.go          # UC: Buscar lances
│   │
│   ├── infra/                               # Infraestrutura
│   │   ├── database/                        # Implementações de repositórios
│   │   │   ├── auction/
│   │   │   │   ├── create_auction.go        # ⭐ Implementação principal
│   │   │   │   └── create_auction_test.go   # ⭐ Testes automatizados
│   │   │   ├── bid/
│   │   │   │   └── create_bid.go            # Repository de lances
│   │   │   └── user/
│   │   │       └── find_user.go             # Repository de usuários
│   │   │
│   │   └── api/
│   │       └── web/
│   │           └── controller/              # Controllers HTTP
│   │               ├── auction_controller/
│   │               │   └── create_auction_controller.go
│   │               └── bid_controller/
│   │                   └── create_bid_controller.go
│   │
│   └── internal_error/                      # Tratamento de erros
│       └── internal_error.go                # Tipos de erro customizados
│
├── .env                                     # Variáveis de ambiente
├── .env.example                             # Exemplo de configuração
├── .gitignore                               # Arquivos ignorados pelo Git
├── docker-compose.yml                       # Orquestração de containers
├── Dockerfile                               # Imagem Docker da aplicação
├── go.mod                                   # Dependências do Go
├── go.sum                                   # Checksums das dependências
├── Makefile                                 # Comandos úteis
│
├── README.md                                # 📖 Documentação principal
├── QUICKSTART.md                            # 🚀 Guia de início rápido
├── IMPLEMENTATION.md                        # 🔧 Detalhes técnicos
├── PROJECT_STRUCTURE.md                     # 📁 Este arquivo
│
├── api-examples.http                        # Exemplos de requisições HTTP
└── test-auto-close.sh                       # Script de teste automatizado

```

## 🎯 Arquivos Principais

### ⭐ Implementação do Fechamento Automático

**`internal/infra/database/auction/create_auction.go`**

Este é o arquivo mais importante do projeto. Contém:

1. **`calculateAuctionDuration()`**
   - Calcula duração do leilão baseado em `AUCTION_DURATION`
   - Retorna valor padrão se não configurado

2. **`getCheckInterval()`**
   - Define intervalo de verificação baseado em `AUCTION_CHECK_INTERVAL`
   - Controla frequência da goroutine

3. **`NewAuctionRepository()`**
   - Inicializa o repository
   - **Inicia a goroutine automaticamente**

4. **`startAuctionExpirationChecker()`** ⭐
   - Goroutine principal
   - Usa `time.Ticker` para execução periódica
   - Chama `closeExpiredAuctions()` a cada intervalo

5. **`closeExpiredAuctions()`** ⭐
   - Busca leilões expirados no MongoDB
   - Atualiza status para `Completed`
   - Usa mutex para thread-safety

6. **`CreateAuction()`**
   - Cria novo leilão
   - Calcula `expires_at`
   - Usa mutex para thread-safety

### ⭐ Testes Automatizados

**`internal/infra/database/auction/create_auction_test.go`**

Contém 9 testes:

1. **`TestAuctionAutomaticClosure`** ⭐ - Teste principal
2. `TestCreateAuction` - Criação básica
3. `TestFindAuctionById` - Busca por ID
4. `TestFindExpiredAuctions` - Busca de expirados
5. `TestUpdateAuctionStatus` - Atualização de status
6. `TestConcurrentAuctionCreation` - Thread-safety
7. `TestCalculateAuctionDuration` - Cálculo de duração
8. `TestCloseExpiredAuctionsDirectly` - Fechamento direto

## 🏗️ Arquitetura

### Clean Architecture

```
┌─────────────────────────────────────────────┐
│              API Layer (Gin)                │
│         (Controllers/Handlers)              │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│           Use Case Layer                    │
│        (Business Logic)                     │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│          Entity Layer                       │
│     (Domain Models + Interfaces)            │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│       Infrastructure Layer                  │
│  (Database, External Services)              │
│                                             │
│  ┌─────────────────────────────────────┐   │
│  │   AuctionRepository                 │   │
│  │   ├── CreateAuction()               │   │
│  │   ├── FindAuction()                 │   │
│  │   └── closeExpiredAuctions() ⭐     │   │
│  │                                     │   │
│  │   Goroutine:                        │   │
│  │   startAuctionExpirationChecker() ⭐│   │
│  └─────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

### Fluxo de Dados

```
HTTP Request
    ↓
Controller (Gin)
    ↓
Use Case (Business Logic)
    ↓
Entity (Domain Model)
    ↓
Repository (Database)
    ↓
MongoDB
```

### Fluxo de Fechamento Automático

```
Application Start
    ↓
NewAuctionRepository()
    ↓
go startAuctionExpirationChecker() ← Goroutine inicia
    ↓
    ├─→ time.Ticker (a cada AUCTION_CHECK_INTERVAL)
    │       ↓
    │   closeExpiredAuctions()
    │       ↓
    │   MongoDB Query (status=Active, expires_at<=now)
    │       ↓
    │   Update status=Completed
    │       ↓
    └─→ Loop infinito (até app encerrar)
```

## 📦 Dependências

### Principais

- **gin-gonic/gin** - Framework web
- **mongo-driver** - Driver MongoDB
- **google/uuid** - Geração de UUIDs
- **godotenv** - Carregamento de .env
- **testify** - Biblioteca de testes

### Go Standard Library

- **sync** - Mutex para concorrência
- **time** - Ticker e duração
- **context** - Controle de timeout
- **log** - Logging

## 🔑 Conceitos Importantes

### 1. Goroutines
```go
go repo.startAuctionExpirationChecker()
```
- Execução concorrente
- Não bloqueia thread principal
- Leve (poucos KB de memória)

### 2. Channels (via Ticker)
```go
ticker := time.NewTicker(interval)
for range ticker.C {
    // Executa a cada intervalo
}
```

### 3. Mutex (Thread-Safety)
```go
ar.mu.Lock()         // Escrita exclusiva
defer ar.mu.Unlock()

ar.mu.RLock()        // Leitura compartilhada
defer ar.mu.RUnlock()
```

### 4. Context
```go
ctx := context.Background()
// Permite timeout e cancelamento
```

### 5. Defer
```go
defer cleanup()  // Executa ao sair da função
```

## 🎨 Padrões de Design

### Repository Pattern
- Abstração do acesso a dados
- Interface define contrato
- Implementação específica do MongoDB

### Dependency Injection
- Repositories injetados nos use cases
- Use cases injetados nos controllers
- Facilita testes e manutenção

### Clean Architecture
- Separação de responsabilidades
- Independência de frameworks
- Testabilidade

## 📝 Convenções de Código

### Nomenclatura
- **Entidades**: `auction_entity.go`
- **Use Cases**: `create_auction_usecase.go`
- **Repositories**: `create_auction.go`
- **Controllers**: `create_auction_controller.go`
- **Testes**: `*_test.go`

### Estrutura de Funções
```go
// 1. Receiver methods
func (ar *AuctionRepository) Method() {}

// 2. Validações primeiro
if err != nil {
    return err
}

// 3. Lógica principal
// ...

// 4. Return
return result, nil
```

### Tratamento de Erros
```go
// Sempre verificar erros
if err != nil {
    log.Printf("Error: %v", err)
    return err
}

// Usar erros customizados
return internal_error.NewBadRequestError("message")
```

## 🧪 Testes

### Estrutura de Teste
```go
func TestFeature(t *testing.T) {
    // 1. Setup
    database, cleanup := setupTestDB(t)
    defer cleanup()
    
    // 2. Execute
    result, err := function()
    
    // 3. Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### Executar Testes
```bash
# Todos os testes
go test ./... -v

# Teste específico
go test ./internal/infra/database/auction -v -run TestAuctionAutomaticClosure

# Com cobertura
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 🐳 Docker

### Estrutura
```
docker-compose.yml
├── mongodb (serviço)
│   └── Porta 27017
└── auction-api (serviço)
    └── Porta 8080
```

### Comandos Úteis
```bash
# Subir
docker-compose up -d

# Logs
docker-compose logs -f auction-api

# Parar
docker-compose down

# Rebuild
docker-compose build --no-cache
```

## 📚 Documentação

- **README.md** - Visão geral e instruções completas
- **QUICKSTART.md** - Início rápido em 3 passos
- **IMPLEMENTATION.md** - Detalhes técnicos da implementação
- **PROJECT_STRUCTURE.md** - Este arquivo

## 🎓 Para Estudar

1. **Goroutines e Concorrência**
   - `internal/infra/database/auction/create_auction.go`
   - Linhas: `startAuctionExpirationChecker()`

2. **Mutex e Thread-Safety**
   - Mesmo arquivo
   - Uso de `sync.RWMutex`

3. **MongoDB com Go**
   - Queries, updates, cursors
   - `configuration/database/mongodb/connection.go`

4. **Clean Architecture**
   - Separação de camadas
   - Fluxo de dependências

5. **Testes em Go**
   - `internal/infra/database/auction/create_auction_test.go`
   - Uso de testify/assert

---

**Dúvidas?** Consulte os arquivos de documentação ou abra uma issue!
