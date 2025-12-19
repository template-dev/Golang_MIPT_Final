SHELL := /bin/sh

PROTO_DIR := proto
PROTO_FILE := proto/ledger/v1/ledger.proto

.PHONY: build test proto migrate-up migrate-down compose-up compose-down logs

build:
	cd gateway && go build ./...
	cd ledger && go build ./...

test:
	go test ./...

proto:
	protoc -I ./proto \
		--go_out=./ledger --go_opt=paths=source_relative \
		--go-grpc_out=./ledger --go-grpc_opt=paths=source_relative \
		$(PROTO_FILE)
	protoc -I ./proto \
		--go_out=./gateway --go_opt=paths=source_relative \
		--go-grpc_out=./gateway --go-grpc_opt=paths=source_relative \
		$(PROTO_FILE)

migrate-up:
	goose -dir ./ledger/migrations postgres $(DATABASE_URL) up

migrate-down:
	goose -dir ./ledger/migrations postgres $(DATABASE_URL) down

compose-up:
	docker compose up -d db
	docker compose up -d ledger gateway

compose-down:
	docker compose down -v

logs:
	docker compose logs -f gateway ledger db
