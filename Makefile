.PHONY: run build test docs lint

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test ./...

lint:
	go vet ./...

docs:
	swag init -g cmd/api/main.go -o docs
