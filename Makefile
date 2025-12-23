.PHONY: all up down db-load app migrate-up migrate-down integration helper-compose-up migrate-helper-compose-up lint

-include .env

all: up 

up: local-compose db-load migrate-up rabbit-load app

local-compose:
	@docker compose -f docker-compose.yaml up -d postgres rabbitmq redis

down:
	@docker compose -f docker-compose.yaml down
	
db-load:
	@until docker exec postgres pg_isready -U ${DB_USER} > /dev/null 2>&1; do sleep 0.5; done

rabbit-load:
	@until docker exec rabbitmq rabbitmqctl status > /dev/null 2>&1; do sleep 0.5; done

app:
	go run ./cmd/chronos/main.go -o app

migrate-up:
	@for i in $$(seq 1 10); do \
		migrate -path ./migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@localhost:5433/chronos-db?sslmode=disable" up && exit 0; \
		echo "Retry $$i/10..."; sleep 1; \
	done; exit 1

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