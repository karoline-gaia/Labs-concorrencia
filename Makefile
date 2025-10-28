.PHONY: help build run test test-coverage docker-build docker-up docker-down docker-logs clean

help: ## Mostra esta mensagem de ajuda
	@echo "Comandos disponíveis:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Compila a aplicação
	@echo "Compilando aplicação..."
	@go build -o auction-api ./cmd/auction

run: ## Executa a aplicação localmente
	@echo "Executando aplicação..."
	@go run cmd/auction/main.go

test: ## Executa todos os testes
	@echo "Executando testes..."
	@go test ./... -v

test-auction: ## Executa apenas os testes de leilão
	@echo "Executando testes de leilão..."
	@go test ./internal/infra/database/auction -v

test-auto-close: ## Executa teste de fechamento automático
	@echo "Executando teste de fechamento automático..."
	@go test ./internal/infra/database/auction -v -run TestAuctionAutomaticClosure

test-coverage: ## Executa testes com cobertura
	@echo "Executando testes com cobertura..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Relatório de cobertura gerado em coverage.html"

docker-build: ## Constrói a imagem Docker
	@echo "Construindo imagem Docker..."
	@docker-compose build

docker-up: ## Inicia os containers
	@echo "Iniciando containers..."
	@docker-compose up -d
	@echo "Containers iniciados!"
	@echo "API disponível em http://localhost:8080"
	@echo "MongoDB disponível em localhost:27017"

docker-down: ## Para os containers
	@echo "Parando containers..."
	@docker-compose down

docker-down-v: ## Para os containers e remove volumes
	@echo "Parando containers e removendo volumes..."
	@docker-compose down -v

docker-logs: ## Mostra os logs dos containers
	@docker-compose logs -f

docker-logs-api: ## Mostra os logs da API
	@docker-compose logs -f auction-api

docker-restart: ## Reinicia os containers
	@echo "Reiniciando containers..."
	@docker-compose restart

clean: ## Remove arquivos gerados
	@echo "Limpando arquivos gerados..."
	@rm -f auction-api
	@rm -f coverage.out coverage.html
	@echo "Limpeza concluída!"

deps: ## Baixa as dependências
	@echo "Baixando dependências..."
	@go mod download
	@go mod tidy

mongo-shell: ## Acessa o shell do MongoDB
	@docker exec -it auction-mongodb mongosh -u admin -p admin

lint: ## Executa o linter
	@echo "Executando linter..."
	@go fmt ./...
	@go vet ./...

install-tools: ## Instala ferramentas de desenvolvimento
	@echo "Instalando ferramentas..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Ferramentas instaladas!"
