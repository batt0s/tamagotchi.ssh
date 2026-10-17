.PHONY: run build tidy fmt test

run:
	go run ./cmd/pet

build:
	go build -o bin/pet ./cmd/pet

tidy:
	go mod tidy

fmt:
	gofmt -w .

test:
	go test ./...
