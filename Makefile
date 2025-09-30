.PHONY: all build test lint clean install

# Binary name
BINARY_NAME=aseprite-mcp

# Build variables
VERSION:=$(shell git describe --tags --always --dirty || echo dev)
BUILD_TIME:=$(shell date -u '+%Y-%m-%d_%H:%M:%S' || echo unknown)
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

all: lint test build

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/aseprite-mcp

test:
	go test -v -race -cover ./...

test-coverage:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html

lint:
	go vet ./...
	go fmt ./...

clean:
	go clean
ifeq ($(OS),Windows_NT)
	-@powershell -NoProfile -Command "if (Test-Path bin) { Remove-Item bin -Recurse -Force }; if (Test-Path dist) { Remove-Item dist -Recurse -Force }; if (Test-Path coverage.txt) { Remove-Item coverage.txt -Force }; if (Test-Path coverage.html) { Remove-Item coverage.html -Force }"
else
	-@rm -rf bin dist coverage.txt coverage.html
endif

install:
	go install $(LDFLAGS) ./cmd/aseprite-mcp

# Development helpers
run:
	go run ./cmd/aseprite-mcp

deps:
	go mod download
	go mod tidy