.PHONY: all up down db-load app migrate-up migrate-down

include .env

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

lint:
	golangci-lint run ./...