.PHONY: dev build run test clean

dev:
	air

# Server
build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

# TUI Client
build-client:
	go build -o bin/vault-tui ./cmd/client

run-client: build-client
	SERVER_URL=http://localhost:8080 ./bin/vault-tui

# Both
build-all: build build-client

test:
	go test ./internal/... -v -count=1

smoke:
	bash scripts/smoke_test.sh

clean:
	rm -rf bin/ tmp/ *.db *.db-journal *.db-wal *.db-shm

# Database
seed:
	go run ./cmd/seed

# Cross-compile for Windows
build-exe:
	GOOS=windows GOARCH=amd64 go build -o bin/classified-vault.exe ./cmd/server

build-client-exe:
	GOOS=windows GOARCH=amd64 go build -o bin/vault-tui.exe ./cmd/client

# Full build (server + client, both platforms)
build-release: build build-exe build-client build-client-exe

# Run with swagger docs regenerated
dev-full:
	swag init -g cmd/server/main.go -o docs/ && air
