# Sistema de Leilão com Fechamento Automático

Este projeto implementa um sistema de leilão online com funcionalidade de fechamento automático usando Go routines e MongoDB.

## 🚀 Funcionalidades

- **Criação de Leilões**: Crie leilões com duração configurável
- **Fechamento Automático**: Leilões são fechados automaticamente quando o tempo expira
- **Sistema de Lances**: Usuários podem fazer lances em leilões ativos
- **Validação de Status**: Sistema valida se o leilão está ativo antes de aceitar lances
- **Concorrência Segura**: Implementação com mutex para operações thread-safe
- **API RESTful**: Interface HTTP para todas as operações

## 🏗️ Arquitetura

O projeto segue os princípios de Clean Architecture:

```
auction-goexpert/
├── cmd/
│   └── auction/
│       └── main.go                 # Ponto de entrada da aplicação
├── configuration/
│   └── database/
│       └── mongodb/
│           └── connection.go       # Configuração do MongoDB
├── internal/
│   ├── entity/                     # Entidades de domínio
│   │   ├── auction_entity.go
│   │   ├── bid_entity.go
│   │   └── user_entity.go
│   ├── usecase/                    # Casos de uso
│   │   ├── auction_usecase/
│   │   └── bid_usecase/
│   ├── infra/
│   │   ├── database/               # Implementação de repositórios
│   │   │   ├── auction/
│   │   │   │   ├── create_auction.go
│   │   │   │   └── create_auction_test.go
│   │   │   ├── bid/
│   │   │   └── user/
│   │   └── api/
│   │       └── web/
│   │           └── controller/     # Controllers HTTP
│   └── internal_error/             # Tratamento de erros
├── .env                            # Variáveis de ambiente
├── docker-compose.yml              # Configuração Docker
├── Dockerfile                      # Imagem Docker
└── README.md                       # Este arquivo
```

## 🔧 Tecnologias Utilizadas

- **Go 1.21**: Linguagem de programação
- **Gin**: Framework web
- **MongoDB**: Banco de dados NoSQL
- **Docker**: Containerização
- **Go Routines**: Concorrência para fechamento automático

## ⚙️ Variáveis de Ambiente

Configure as seguintes variáveis no arquivo `.env`:

```env
MONGODB_URI=mongodb://admin:admin@mongodb:27017
MONGODB_DATABASE=auctions
AUCTION_DURATION=300           # Duração do leilão em segundos (padrão: 5 minutos)
AUCTION_CHECK_INTERVAL=10      # Intervalo de verificação em segundos (padrão: 10 segundos)
```

### Descrição das Variáveis

- **MONGODB_URI**: String de conexão com o MongoDB
- **MONGODB_DATABASE**: Nome do banco de dados
- **AUCTION_DURATION**: Tempo de duração de cada leilão em segundos
- **AUCTION_CHECK_INTERVAL**: Intervalo em que a goroutine verifica leilões expirados

## 🐳 Como Executar com Docker

### Pré-requisitos

- Docker
- Docker Compose

### Passos

1. **Clone o repositório** (ou navegue até o diretório do projeto)

```bash
cd auction-goexpert
```

2. **Inicie os containers**

```bash
docker-compose up -d
```

Este comando irá:
- Criar um container MongoDB na porta 27017
- Compilar e executar a aplicação na porta 8080
- Configurar a rede entre os containers

3. **Verifique os logs**

```bash
docker-compose logs -f auction-api
```

4. **Pare os containers**

```bash
docker-compose down
```

5. **Pare e remova os volumes (limpa o banco de dados)**

```bash
docker-compose down -v
```

## 💻 Como Executar em Desenvolvimento

### Pré-requisitos

- Go 1.21 ou superior
- MongoDB rodando localmente ou via Docker

### Passos

1. **Inicie o MongoDB** (se não estiver usando Docker Compose)

```bash
docker run -d -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=admin \
  mongo:7.0
```

2. **Configure as variáveis de ambiente**

Copie o arquivo `.env.example` para `.env` e ajuste conforme necessário:

```bash
cp .env.example .env
```

3. **Instale as dependências**

```bash
go mod download
```

4. **Execute a aplicação**

```bash
go run cmd/auction/main.go
```

A API estará disponível em `http://localhost:8080`

## 🧪 Executar Testes

### Todos os testes

```bash
go test ./... -v
```

### Testes específicos do fechamento automático

```bash
go test ./internal/infra/database/auction -v -run TestAuctionAutomaticClosure
```

### Teste de concorrência

```bash
go test ./internal/infra/database/auction -v -run TestConcurrentAuctionCreation
```

### Cobertura de testes

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 📡 Endpoints da API

### Leilões

#### Criar Leilão

```http
POST /auction
Content-Type: application/json

{
  "product_name": "iPhone 13",
  "category": "Electronics",
  "description": "Brand new iPhone 13 with 128GB storage",
  "condition": 0
}
```

**Condições:**
- `0`: Novo
- `1`: Usado
- `2`: Recondicionado

**Resposta:**

```json
{
  "id": "uuid",
  "product_name": "iPhone 13",
  "category": "Electronics",
  "description": "Brand new iPhone 13 with 128GB storage",
  "condition": 0,
  "status": 0,
  "timestamp": "2024-01-15T10:00:00Z",
  "expires_at": "2024-01-15T10:05:00Z"
}
```

#### Buscar Leilão por ID

```http
GET /auction/:auctionId
```

#### Listar Leilões

```http
GET /auction?status=0&category=Electronics&productName=iPhone
```

**Parâmetros de Query (opcionais):**
- `status`: 0 (Ativo) ou 1 (Completo)
- `category`: Categoria do produto
- `productName`: Nome do produto (busca parcial)

### Lances

#### Criar Lance

```http
POST /bid
Content-Type: application/json

{
  "user_id": "user-uuid",
  "auction_id": "auction-uuid",
  "amount": 1500.00
}
```

**Validações:**
- O leilão deve existir
- O leilão deve estar ativo (status = 0)
- O leilão não pode estar expirado
- O valor deve ser maior que zero

#### Buscar Lances de um Leilão

```http
GET /bid/auction/:auctionId
```

#### Buscar Lance Vencedor

```http
GET /bid/auction/:auctionId/winner
```

Retorna o lance com maior valor para o leilão especificado.

## 🔄 Funcionamento do Fechamento Automático

### Implementação

O fechamento automático é implementado no arquivo `internal/infra/database/auction/create_auction.go` através de:

1. **Cálculo de Duração**: A função `calculateAuctionDuration()` lê a variável de ambiente `AUCTION_DURATION` e define o tempo de expiração do leilão.

2. **Go Routine de Verificação**: Quando o `AuctionRepository` é inicializado, uma goroutine é iniciada automaticamente através do método `startAuctionExpirationChecker()`.

3. **Ticker Periódico**: A goroutine usa um `time.Ticker` para verificar periodicamente (baseado em `AUCTION_CHECK_INTERVAL`) se existem leilões expirados.

4. **Fechamento Automático**: O método `closeExpiredAuctions()` busca todos os leilões ativos com `expires_at <= now` e atualiza seu status para `Completed`.

5. **Thread Safety**: Utiliza `sync.RWMutex` para garantir operações seguras em ambiente concorrente:
   - `RLock/RUnlock`: Para operações de leitura
   - `Lock/Unlock`: Para operações de escrita

### Fluxo de Execução

```
1. Aplicação inicia
   ↓
2. AuctionRepository é criado
   ↓
3. Goroutine de verificação inicia automaticamente
   ↓
4. A cada AUCTION_CHECK_INTERVAL segundos:
   - Busca leilões com status=Active e expires_at <= now
   - Atualiza status para Completed
   - Registra log da operação
   ↓
5. Continua executando até a aplicação encerrar
```

## 🧪 Testes Implementados

### TestAuctionAutomaticClosure

Teste principal que valida o fechamento automático:

1. Cria um leilão com duração de 3 segundos
2. Configura intervalo de verificação de 1 segundo
3. Verifica que o leilão está ativo inicialmente
4. Aguarda 5 segundos (tempo de expiração + margem)
5. Verifica que o status foi alterado para Completed

### Outros Testes

- **TestCreateAuction**: Valida criação de leilão
- **TestFindAuctionById**: Valida busca por ID
- **TestFindExpiredAuctions**: Valida busca de leilões expirados
- **TestUpdateAuctionStatus**: Valida atualização de status
- **TestConcurrentAuctionCreation**: Valida criação concorrente (thread safety)
- **TestCalculateAuctionDuration**: Valida cálculo de duração
- **TestCloseExpiredAuctionsDirectly**: Valida fechamento direto

## 📊 Exemplos de Uso

### Exemplo Completo

```bash
# 1. Criar um leilão
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "MacBook Pro M1",
    "category": "Electronics",
    "description": "MacBook Pro 2021 with M1 chip, 16GB RAM, 512GB SSD",
    "condition": 0
  }'

# Resposta: {"id":"abc-123",...,"expires_at":"2024-01-15T10:05:00Z"}

# 2. Fazer um lance
curl -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "auction_id": "abc-123",
    "amount": 2500.00
  }'

# 3. Fazer outro lance
curl -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-456",
    "auction_id": "abc-123",
    "amount": 2800.00
  }'

# 4. Buscar o lance vencedor
curl http://localhost:8080/bid/auction/abc-123/winner

# 5. Aguardar o leilão expirar (5 minutos por padrão)
# O sistema fechará automaticamente

# 6. Verificar que o leilão foi fechado
curl http://localhost:8080/auction/abc-123
# Resposta: {"id":"abc-123",...,"status":1}

# 7. Tentar fazer um lance após expiração (deve falhar)
curl -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-789",
    "auction_id": "abc-123",
    "amount": 3000.00
  }'
# Resposta: {"error":"auction has expired"}
```

## 🔍 Monitoramento

### Logs

A aplicação gera logs detalhados:

```
2024/01/15 10:00:00 Connected to MongoDB successfully
2024/01/15 10:00:00 Auction expiration checker started with interval: 10s
2024/01/15 10:00:05 Auction created successfully: abc-123, expires at: 2024-01-15T10:05:05Z
2024/01/15 10:05:10 Auction abc-123 closed automatically (expired at: 2024-01-15T10:05:05Z)
2024/01/15 10:05:10 Closed 1 expired auction(s)
```

### Verificar Status dos Containers

```bash
docker-compose ps
```

### Acessar MongoDB

```bash
docker exec -it auction-mongodb mongosh -u admin -p admin
```

## 🛠️ Troubleshooting

### Problema: Aplicação não conecta ao MongoDB

**Solução**: Verifique se o MongoDB está rodando:

```bash
docker-compose ps mongodb
```

### Problema: Leilões não estão fechando automaticamente

**Solução**: Verifique os logs e as variáveis de ambiente:

```bash
docker-compose logs auction-api | grep "expiration checker"
```

### Problema: Testes falhando

**Solução**: Certifique-se de que o MongoDB está rodando na porta 27017:

```bash
docker run -d -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=admin \
  mongo:7.0
```

## 📝 Notas Importantes

1. **Concorrência**: O sistema usa `sync.RWMutex` para garantir operações thread-safe
2. **Goroutines**: Uma única goroutine gerencia o fechamento de todos os leilões
3. **Performance**: O intervalo de verificação pode ser ajustado conforme a carga
4. **Escalabilidade**: Para produção, considere usar um sistema de filas (RabbitMQ, Kafka)

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto é livre para uso educacional.

## ✨ Autor

Desenvolvido como parte do desafio Go Expert - Full Cycle

---

**Dúvidas?** Abra uma issue no repositório!
