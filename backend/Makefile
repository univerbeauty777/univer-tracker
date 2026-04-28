.PHONY: help dev build run test lint fmt tidy migrate-up migrate-down docker-build

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Run API in development mode with hot reload
	@command -v air >/dev/null 2>&1 || go install github.com/air-verse/air@latest
	air -c .air.toml

build: ## Build API and worker binaries
	@mkdir -p bin
	go build -trimpath -ldflags="-s -w" -o bin/api ./cmd/api
	go build -trimpath -ldflags="-s -w" -o bin/worker ./cmd/worker

run: ## Run API binary
	./bin/api

worker: ## Run worker binary
	./bin/worker

test: ## Run tests
	go test -race -cover ./...

lint: ## Run linter
	@command -v golangci-lint >/dev/null 2>&1 || (echo "Install golangci-lint: https://golangci-lint.run" && exit 1)
	golangci-lint run ./...

fmt: ## Format code
	gofmt -s -w .
	@command -v goimports >/dev/null 2>&1 && goimports -w . || true

tidy: ## Tidy dependencies
	go mod tidy

migrate-up: ## Run migrations
	@command -v migrate >/dev/null 2>&1 || (echo "Install golang-migrate" && exit 1)
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down: ## Rollback last migration
	migrate -path migrations -database "$(DATABASE_URL)" down 1

docker-build: ## Build Docker image
	docker build -t univer-tracker-backend:latest .
