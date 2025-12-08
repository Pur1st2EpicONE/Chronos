.PHONY: all up down db-load app migrate-up migrate-down

include .env

all: up 

up: local-compose db-load migrate-up app

local-compose:
	@docker compose -f docker-compose.yaml up -d postgres

down:
	@docker compose -f docker-compose.yaml down
	
db-load:
	@until docker exec postgres pg_isready -U ${DB_USER} > /dev/null 2>&1; do sleep 0.5; done

app:
	go run ./main.go -o app

migrate-up:
	@migrate -path ./migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@localhost:5433/chronos-db?sslmode=disable" up

migrate-down:
	@migrate -path ./migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@localhost:5433/chronos-db?sslmode=disable" down

lint:
	golangci-lint run ./...