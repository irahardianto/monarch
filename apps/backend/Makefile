.PHONY: test build generate

generate:
	sqlc generate

test:
	go test ./...

build:
	go build -o bin/monarch ./cmd/monarch

run:
	go run ./cmd/monarch
