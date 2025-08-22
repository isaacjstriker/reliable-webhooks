APP_NAME=webhooks
GO_FILES=$(shell find . -name "*.go")
DB_URL=postgres://postgres:postgres@localhost:5432/webhooks?sslmode=disable


.PHONY: tidy run dev migrate-up migrate-down build docker-up docker-down test lint

tidy:
	go mod tidy

run:
	go run ./cmd/api

dev:
	REFRESH_INTERNAL=2s air || go run ./cmd/api

migrate-up:
	psql "$(DB_URL)" -f migrations/0001_init.sql

migrate-down:
	psql "$(DB_URL)" -c "DROP TABLE IF EXISTS events CASCADE; DROP TYPE IF EXISTS event_status;"

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