.PHONY: run test lint build clean install-tools ngrok

run:
	go run main.go

test:
	go test ./...

lint:
	golangci-lint run

build:
	go build -o socgo main.go

templ:
	templ generate

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/air-verse/air@latest