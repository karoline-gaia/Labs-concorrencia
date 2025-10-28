# 🚀 Guia de Início Rápido

## Iniciar o Projeto em 3 Passos

### 1️⃣ Subir os containers

```bash
cd auction-goexpert
docker-compose up -d
```

Aguarde alguns segundos para os serviços iniciarem.

### 2️⃣ Verificar se está funcionando

```bash
# Ver logs da aplicação
docker-compose logs -f auction-api

# Você deve ver:
# ✓ Connected to MongoDB successfully
# ✓ Auction expiration checker started with interval: 10s
# ✓ Server starting on port 8080...
```

### 3️⃣ Testar a API

```bash
# Criar um leilão
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "iPhone 13",
    "category": "Electronics",
    "description": "Brand new iPhone 13 with 128GB",
    "condition": 0
  }'

# Resposta esperada:
# {
#   "id": "abc-123...",
#   "product_name": "iPhone 13",
#   "status": 0,
#   "expires_at": "2024-01-15T10:05:00Z"
# }
```

## 🧪 Testar Fechamento Automático

### Opção A: Teste Rápido (30 segundos)

```bash
# 1. Configure duração curta
echo "MONGODB_URI=mongodb://admin:admin@mongodb:27017
MONGODB_DATABASE=auctions
AUCTION_DURATION=20
AUCTION_CHECK_INTERVAL=5" > .env

# 2. Reinicie a aplicação
docker-compose restart auction-api

# 3. Crie um leilão e aguarde 25 segundos
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "Teste Rápido",
    "category": "Test",
    "description": "Leilão que expira em 20 segundos",
    "condition": 0
  }'

# Salve o ID retornado e aguarde 25 segundos...

# 4. Verifique o status (deve ser 1 = Completo)
curl http://localhost:8080/auction/SEU_ID_AQUI
```

### Opção B: Script Automatizado

```bash
# Execute o script de teste
./test-auto-close.sh

# O script irá:
# ✓ Criar um leilão
# ✓ Fazer alguns lances
# ✓ Aguardar a expiração
# ✓ Verificar o fechamento automático
# ✓ Tentar fazer lance após fechamento (deve falhar)
```

### Opção C: Testes Unitários

```bash
# Executar todos os testes
docker-compose exec auction-api go test ./... -v

# Executar apenas teste de fechamento automático
docker-compose exec auction-api go test ./internal/infra/database/auction -v -run TestAuctionAutomaticClosure

# Ver cobertura
docker-compose exec auction-api go test ./... -coverprofile=coverage.out
docker-compose exec auction-api go tool cover -func=coverage.out
```

## 📡 Endpoints Principais

### Criar Leilão
```bash
POST /auction
```

### Listar Leilões Ativos
```bash
GET /auction?status=0
```

### Criar Lance
```bash
POST /bid
```

### Ver Lance Vencedor
```bash
GET /bid/auction/:auctionId/winner
```

## 🔍 Monitorar Logs

```bash
# Logs em tempo real
docker-compose logs -f auction-api

# Buscar por fechamentos automáticos
docker-compose logs auction-api | grep "closed automatically"

# Ver todos os leilões criados
docker-compose logs auction-api | grep "Auction created"
```

## 🛑 Parar o Projeto

```bash
# Parar containers
docker-compose down

# Parar e limpar dados
docker-compose down -v
```

## 🐛 Troubleshooting

### Problema: Porta 8080 já em uso
```bash
# Altere a porta no docker-compose.yml
ports:
  - "8081:8080"  # Use 8081 no host
```

### Problema: MongoDB não conecta
```bash
# Verifique se o MongoDB está rodando
docker-compose ps mongodb

# Reinicie o MongoDB
docker-compose restart mongodb
```

### Problema: Aplicação não inicia
```bash
# Veja os logs de erro
docker-compose logs auction-api

# Reconstrua a imagem
docker-compose build --no-cache
docker-compose up -d
```

## 📊 Exemplo Completo de Uso

```bash
# 1. Criar leilão
AUCTION=$(curl -s -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "MacBook Pro",
    "category": "Electronics",
    "description": "MacBook Pro M1 16GB",
    "condition": 0
  }')

AUCTION_ID=$(echo $AUCTION | jq -r '.id')
echo "Leilão criado: $AUCTION_ID"

# 2. Fazer lances
curl -s -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user-1\",
    \"auction_id\": \"$AUCTION_ID\",
    \"amount\": 2000.00
  }" | jq '.'

curl -s -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user-2\",
    \"auction_id\": \"$AUCTION_ID\",
    \"amount\": 2500.00
  }" | jq '.'

# 3. Ver lance vencedor
curl -s http://localhost:8080/bid/auction/$AUCTION_ID/winner | jq '.'

# 4. Ver status do leilão
curl -s http://localhost:8080/auction/$AUCTION_ID | jq '.'

# 5. Aguardar expiração (300s por padrão)
# Depois verificar novamente o status
```

## 🎯 Próximos Passos

1. ✅ Explore os endpoints na API (veja `api-examples.http`)
2. ✅ Leia a documentação completa (`README.md`)
3. ✅ Entenda a implementação (`IMPLEMENTATION.md`)
4. ✅ Execute os testes automatizados
5. ✅ Customize as variáveis de ambiente

## 💡 Dicas

- Use `jq` para formatar JSON: `curl ... | jq '.'`
- Configure durações curtas para testes rápidos
- Monitore os logs para ver o fechamento automático
- Use o Makefile para comandos comuns: `make help`

---

**Pronto para começar?** Execute `docker-compose up -d` e comece a testar! 🚀
