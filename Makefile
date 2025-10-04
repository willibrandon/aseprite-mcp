.PHONY: all build release test test-nocache test-integration test-integration-nocache bench lint clean install

# Binary name
BINARY_NAME=aseprite-mcp
ifeq ($(OS),Windows_NT)
	BINARY_NAME := $(BINARY_NAME).exe
endif

# Build variables for releases
VERSION:=$(shell git describe --tags --always --dirty || echo dev)
ifeq ($(OS),Windows_NT)
	BUILD_TIME:=$(shell powershell -NoProfile -Command "[System.DateTime]::UtcNow.ToString('yyyy-MM-dd_HH:mm:ss')" || echo unknown)
else
	BUILD_TIME:=$(shell date -u '+%Y-%m-%d_%H:%M:%S' || echo unknown)
endif
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

all: lint test build

build:
	go build -o bin/$(BINARY_NAME) ./cmd/aseprite-mcp

release:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/aseprite-mcp

test:
	go test -v -race -cover ./...

test-nocache:
	go test -v -race -cover -count=1 ./...

test-integration:
	go test -tags=integration -v ./...

test-integration-nocache:
	go test -tags=integration -v -count=1 ./...

test-coverage:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html

bench:
	go test -tags=integration -bench=. -benchmem ./pkg/aseprite ./pkg/tools

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