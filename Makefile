.PHONY: run test lint build clean install-tools

run:
	go run ./cmd/socgo

test:
	go test ./...

lint:
	golangci-lint run

build:
	go build -o bin/socgo ./cmd/socgo

clean:
	rm -rf bin/

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/air-verse/air@latest