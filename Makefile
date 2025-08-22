include .env

DB_URL=postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
PSQL_BASE=postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/postgres?sslmode=disable

APP_NAME=webhooks

.PHONY: tidy run dev migrate-up migrate-down create-db build docker-up docker-down test lint

tidy:
	go mod tidy

create-db:
	psql "$(PSQL_BASE)" -tc "SELECT 1 FROM pg_database WHERE datname='$(DB_NAME)';" | grep -q 1 || psql "$(PSQL_BASE)" -c "CREATE DATABASE $(DB_NAME);"

migrate-up: create-db
	psql "$(DB_URL)" -f migrations/0001_init.sql

migrate-down:
	psql "$(DB_URL)" -c "DROP TABLE IF EXISTS events CASCADE; DROP TYPE IF EXISTS event_status;"

run:
	go run ./cmd/api

dev:
	REFRESH_INTERVAL=2s air || go run ./cmd/api

build:
	go build -o bin/$(APP_NAME) ./cmd/api

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down -v

test:
	go test ./...

lint:
	go vet ./...