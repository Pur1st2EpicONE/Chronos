.PHONY: all up down db-load app migrate-up migrate-down integration helper-compose-up migrate-helper-compose-up lint

-include .env.example .env

all: full

full:
	@if [ ! -f .env ] && [ ! -f .env.example ]; then \
		echo "Missing environment file: .env or .env.example is required."; \
		exit 1; \
	fi
	@if [ ! -f .env ]; then cat .env.example > .env; fi
	@if [ ! -f config.yaml ]; then cp ./configs/config.full.yaml ./config.yaml; fi
	@if [ ! -f docker-compose.yaml ]; then cp ./deployments/docker-compose.full.yaml ./docker-compose.yaml; fi
	@if [ ! -f Dockerfile ]; then cp ./deployments/Dockerfile ./Dockerfile; fi
	@docker-compose build --no-cache
	@docker-compose up -d postgres rabbitmq redis
	@$(MAKE) -s db-load
	@$(MAKE) -s rabbit-load
	@docker-compose up -d app 2>&1 | grep -v "is up-to-date"

local: local-compose db-load migrate-up rabbit-load app

local-compose:
	@cat .env.example > .env
	@cp ./configs/config.dev.yaml ./config.yaml
	@cp ./deployments/docker-compose.dev.yaml ./docker-compose.yaml
	@docker compose up -d postgres rabbitmq redis

down:
	@docker compose down 2>/dev/null || true
	@rm -f config.yaml
	@rm -f Dockerfile
	@rm -f docker-compose.yaml
	@rm -f .env
	
db-load:
	@until docker exec postgres pg_isready -U ${DB_USER} > /dev/null 2>&1; do sleep 0.5; done

migrate-up:
	@for i in $$(seq 1 10); do \
		migrate -path ./migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@localhost:5433/chronos-db?sslmode=disable" up && exit 0; \
		echo "Retry $$i/10..."; sleep 1; \
	done; exit 1

rabbit-load:
	@until docker exec rabbitmq rabbitmqctl status > /dev/null 2>&1; do sleep 0.5; done

app:
	@bash -c 'trap "exit 0" INT; go run ./cmd/chronos/main.go'

migrate-down:
	@migrate -path ./migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@localhost:5433/chronos-db?sslmode=disable" down

test:
	@go test -cover ./internal/handler/v1/...
	@go test -cover ./internal/service/impl/...
	@$(MAKE) integration --no-print-directory

integration: migrate-helper-compose-up
	@go test ./internal/repository/postgres -cover
	@docker compose -f docker-compose.yaml stop postgres-test > /dev/null 2>&1
	@docker compose -f docker-compose.yaml rm -f postgres-test > /dev/null 2>&1

helper-compose-up:
	@docker compose -f docker-compose.yaml up -d postgres-test > /dev/null 2>&1

helper-db-load:
	@until docker exec postgres-test pg_isready -U ${DB_USER} > /dev/null 2>&1; do sleep 0.5; done

migrate-helper-compose-up: helper-compose-up helper-db-load
	@for i in $$(seq 1 10); do \
		migrate -path ./migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@localhost:5434/chronos_test?sslmode=disable" up > /dev/null 2>&1 && exit 0; sleep 1; \
	done; exit 1

lint:
	golangci-lint run ./...