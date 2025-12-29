.PHONY: build test install clean fmt lint deps help

# Build variables
BINARY_NAME=nocturnal
VERSION?=1.0.0-go
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

## build: Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

## test: Run tests
test:
	go test -v -race ./...

## install: Install to ~/.local/bin
install:
	go build $(LDFLAGS) -o $(BINARY_NAME) .
	cp $(BINARY_NAME) ~/.local/bin/
	rm $(BINARY_NAME)

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	go clean

## fmt: Format code
fmt:
	go fmt ./...

## lint: Run linter
lint:
	go vet ./...

## deps: Download dependencies
deps:
	go mod download
	go mod tidy

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/ /'
