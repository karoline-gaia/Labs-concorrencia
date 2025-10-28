#!/bin/bash

echo "=========================================="
echo "Teste de Fechamento Automático de Leilão"
echo "=========================================="
echo ""

# Cores para output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# URL da API
API_URL="http://localhost:8080"

echo -e "${BLUE}1. Criando um leilão...${NC}"
RESPONSE=$(curl -s -X POST "$API_URL/auction" \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "Teste Auto-Close",
    "category": "Test",
    "description": "Produto de teste para fechamento automático",
    "condition": 0
  }')

echo "$RESPONSE" | jq '.'

# Extrai o ID do leilão
AUCTION_ID=$(echo "$RESPONSE" | jq -r '.id')

if [ "$AUCTION_ID" == "null" ] || [ -z "$AUCTION_ID" ]; then
    echo -e "${RED}Erro ao criar leilão!${NC}"
    exit 1
fi

echo -e "${GREEN}Leilão criado com ID: $AUCTION_ID${NC}"
echo ""

echo -e "${BLUE}2. Verificando status inicial do leilão...${NC}"
curl -s "$API_URL/auction/$AUCTION_ID" | jq '.'
echo ""

echo -e "${BLUE}3. Criando alguns lances...${NC}"
curl -s -X POST "$API_URL/bid" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user-001\",
    \"auction_id\": \"$AUCTION_ID\",
    \"amount\": 100.00
  }" | jq '.'

curl -s -X POST "$API_URL/bid" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user-002\",
    \"auction_id\": \"$AUCTION_ID\",
    \"amount\": 150.00
  }" | jq '.'

curl -s -X POST "$API_URL/bid" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user-003\",
    \"auction_id\": \"$AUCTION_ID\",
    \"amount\": 200.00
  }" | jq '.'

echo ""
echo -e "${GREEN}3 lances criados com sucesso!${NC}"
echo ""

echo -e "${BLUE}4. Buscando lance vencedor atual...${NC}"
curl -s "$API_URL/bid/auction/$AUCTION_ID/winner" | jq '.'
echo ""

# Verifica a duração configurada
DURATION=${AUCTION_DURATION:-300}
echo -e "${BLUE}5. Aguardando o leilão expirar (duração configurada: ${DURATION}s)...${NC}"
echo "Você pode acompanhar os logs da aplicação em outro terminal:"
echo "  docker-compose logs -f auction-api"
echo ""
echo "Pressione Ctrl+C para cancelar ou aguarde..."
echo ""

# Aguarda a duração + margem de segurança
WAIT_TIME=$((DURATION + 15))
for i in $(seq 1 $WAIT_TIME); do
    echo -ne "Aguardando... $i/${WAIT_TIME}s\r"
    sleep 1
done
echo ""

echo -e "${BLUE}6. Verificando se o leilão foi fechado automaticamente...${NC}"
FINAL_STATUS=$(curl -s "$API_URL/auction/$AUCTION_ID")
echo "$FINAL_STATUS" | jq '.'
echo ""

STATUS=$(echo "$FINAL_STATUS" | jq -r '.status')
if [ "$STATUS" == "1" ]; then
    echo -e "${GREEN}✓ SUCESSO! O leilão foi fechado automaticamente!${NC}"
    echo -e "${GREEN}  Status mudou de 0 (Ativo) para 1 (Completo)${NC}"
else
    echo -e "${RED}✗ FALHA! O leilão ainda está ativo (status: $STATUS)${NC}"
fi
echo ""

echo -e "${BLUE}7. Tentando fazer um lance após o fechamento (deve falhar)...${NC}"
ERROR_RESPONSE=$(curl -s -X POST "$API_URL/bid" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user-004\",
    \"auction_id\": \"$AUCTION_ID\",
    \"amount\": 300.00
  }")

echo "$ERROR_RESPONSE" | jq '.'

if echo "$ERROR_RESPONSE" | jq -e '.error' > /dev/null; then
    echo -e "${GREEN}✓ Correto! Lance rejeitado após fechamento do leilão${NC}"
else
    echo -e "${RED}✗ Atenção! Lance foi aceito mesmo após fechamento${NC}"
fi

echo ""
echo "=========================================="
echo "Teste concluído!"
echo "=========================================="
