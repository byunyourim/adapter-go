.PHONY: build run test lint tidy migrate-up migrate-down generate

build:
	go build -o bin/adapter ./cmd/adapter

run:
	go run ./cmd/adapter

test:
	go test ./...

lint:
	golangci-lint run

tidy:
	go mod tidy

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

# Ent 코드 생성 (ent/schema/ 정의 → 클라이언트/쿼리 코드)
generate:
	go generate ./...
