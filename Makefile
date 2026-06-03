.PHONY: dev build run test clean

dev:
	air

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

test:
	go test ./internal/... -v -count=1

clean:
	rm -rf bin/ tmp/ *.db *.db-journal *.db-wal *.db-shm

# Cross-compile for Windows
build-exe:
	GOOS=windows GOARCH=amd64 go build -o bin/classified-vault.exe ./cmd/server

# Full build (server + Windows client)
build-all: build build-exe

# Run with swagger docs regenerated
dev-full:
	swag init -g cmd/server/main.go -o docs/ && air
