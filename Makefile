include .env

.PHONY: help setup dev dev-up dev-down build test clean docker docker-down install-deps run

BINARY_NAME=main
DOCKER_COMPOSE_FILE=docker-compose.yml
DOCKER_COMPOSE_DEV_FILE=docker-compose.dev.yml

MIGRATE_CMD=migrate
MIGRATE_PATH=migrations
DB_URL=mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)?multiStatements=true

# Instalar dependências
install-deps: ## Instala as dependências do Go
	@echo "📦 Instalando dependências..."
	@go mod download
	@go mod tidy

# Instalar ferramentas de desenvolvimento
install-tools:
	@echo "🛠️  Instalando ferramentas de desenvolvimento..."
	@go install github.com/cespare/reflex@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Ferramentas instaladas"

# Desenvolvimento
dev-up: ## Sobe os serviços de desenvolvimento (PostgreSQL e Redis)
	@echo "🐳 Subindo serviços de desenvolvimento..."
	@docker compose up postgres keycloak -d
	@echo "✅ Serviços rodando:"

dev-down: ## Para os serviços de desenvolvimento
	@echo "🛑 Parando serviços de desenvolvimento..."
	@docker compose down postgres keycloak

# Build
build: ## Compila a aplicação
	@echo "🔨 Compilando aplicação..."
	@go build -o bin/$(BINARY_NAME) cmd/api/main.go
	@echo "✅ Binário criado: bin/$(BINARY_NAME)"

# Executar aplicação compilada
run: build ## Executa a aplicação compilada
	@echo "🚀 Executando aplicação..."
	@./bin/$(BINARY_NAME)

run-dev: ## Inicia o servidor com hot reload
	@echo "🔥 Iniciando servidor com hot reload..."
	@echo "📍 Servidor rodará em: http://localhost:3030"
	@echo "🔄 Arquivos monitorados para reload automático"
	@echo "Press Ctrl+C to stop"
	@reflex -c reflex.conf

# Testes
test: ## Executa os testes
	@echo "🧪 Executando testes..."
	@go test ./...

test-v: ## Executa os testes
	@echo "🧪 Executando testes..."
	@go test -v ./...

test-cover: ## Executa testes com cobertura
	@echo "🧪 Executando testes com cobertura..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Relatório de cobertura gerado: coverage.html"

# Linting
lint: ## Executa o linter
	@echo "🔍 Executando linter..."
	@golangci-lint run

# Limpeza
clean: ## Remove arquivos temporários e binários
	@echo "🧹 Limpando arquivos temporários..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@docker system prune -f
	@echo "✅ Limpeza concluída"

# Database
migrate-create:
	@echo "🔨 Criando migração..."
	@$(MIGRATE_CMD) create -dir=$(MIGRATE_PATH) -ext=sql -seq $(NAME)

migrate:
	@echo "📊 Executando migrações..."
	@$(MIGRATE_CMD) -path $(MIGRATE_PATH) -database "$(DB_URL)" up
	@echo "✅ Migração concluída"

rollback:
	@echo "⏪ Executando rollback das migrações..."
	@$(MIGRATE_CMD) -path $(MIGRATE_PATH) -database "$(DB_URL)" down 1
	@echo "✅ Rollback concluído"