.PHONY: all build test lint clean install

# Binary name
BINARY_NAME=aseprite-mcp

# Build variables
VERSION?=$(shell git describe --tags --always --dirty 2>nul || echo "dev")
BUILD_TIME=$(shell powershell -Command "Get-Date -Format 'yyyy-MM-dd_HH:mm:ss'")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

all: lint test build

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME).exe ./cmd/aseprite-mcp

test:
	go test -v -race -cover ./...

test-coverage:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html

lint:
	go vet ./...
	go fmt ./...

clean:
	if exist bin rmdir /s /q bin
	if exist dist rmdir /s /q dist
	if exist coverage.txt del coverage.txt
	if exist coverage.html del coverage.html
	go clean

install:
	go install $(LDFLAGS) ./cmd/aseprite-mcp

# Development helpers
run:
	go run ./cmd/aseprite-mcp

deps:
	go mod download
	go mod tidy