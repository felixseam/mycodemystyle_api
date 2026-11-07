# config
-include .env
export

APP_NAME   ?= rerng-api
BIN_DIR    ?= tmp
BIN        ?= $(BIN_DIR)/main
MIGRATIONS ?= migrations
DATA_DIR   ?= .data
COMPOSE    ?= docker compose -f deploy/docker-compose.yml

# DATABASE_URL comes from .env
GOOSE = goose -dir $(MIGRATIONS) postgres "$(PG_DATABASE_URL)"

.DEFAULT_GOAL := help
.PHONY: help run dev build test cover fmt vet lint tidy check clean \
        tarantool-up tarantool-clean \
        migrate-new migrate-up migrate-down migrate-status migrate-reset \
        docker-up docker-down docker-logs tools

# help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n\n"} \
		/^[a-zA-Z0-9_-]+:.*##/ {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# app
run: ## Run the API once (go run)
	go run .

dev: ## Run with live reload (air)
	air

build: ## Build binary to tmp/main
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN) .

# quality
test: ## Run all tests
	go test ./... -race -count=1

cover: ## Tests with coverage report
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -n 1

fmt: ## Format code
	gofmt -s -w .

vet: ## Static checks
	go vet ./...

lint: ## Lint (needs golangci-lint)
	golangci-lint run ./...

tidy: ## Clean up go.mod / go.sum
	go mod tidy

check: fmt vet test ## Format, vet and test before commit

# tarantool
tarantool-up: ## Start Tarantool (data in .data/tarantool)
	@mkdir -p $(DATA_DIR)/tarantool
	tarantool deploy/tarantool/init.lua

tarantool-clean: ## Delete local Tarantool data (cache only!)
	rm -rf $(DATA_DIR)/tarantool

# database migrations
migrate-new: ## Create migration: make migrate-new name=create_movies
	@test -n "$(name)" || (echo "usage: make migrate-new name=create_movies" && exit 1)
	goose -dir $(MIGRATIONS) -s create $(name) sql

migrate-up: ## Apply all pending migrations
	$(GOOSE) up

migrate-down: ## Roll back the last migration
	$(GOOSE) down

migrate-status: ## Show migration status
	$(GOOSE) status

migrate-reset: ## Roll back ALL migrations (dev only)
	@read -p "This wipes the schema. Type 'yes' to continue: " ans; \
	[ "$$ans" = "yes" ] && $(GOOSE) reset || echo "cancelled"

# docker
docker-up: ## Start services from deploy/docker-compose.yml
	$(COMPOSE) up -d

docker-down: ## Stop services
	$(COMPOSE) down

docker-logs: ## Follow service logs
	$(COMPOSE) logs -f

# tools / cleanup
tools: ## Install dev tools (air, goose)
	go install github.com/air-verse/air@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest

clean: ## Remove build output and coverage files
	rm -rf $(BIN_DIR) coverage.out