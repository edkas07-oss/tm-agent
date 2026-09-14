# Makefile for tm-agent (Unified Cross-Platform Event Collector Daemon)

BINARY_NAME=tm-agent
VERSION=$(shell cat VERSION 2>/dev/null || echo "0.1.0")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-s -w -X 'github.com/eddywiyatno/tm-agent/internal/buildinfo.Version=$(VERSION)' \
           -X 'github.com/eddywiyatno/tm-agent/internal/buildinfo.GitCommit=$(GIT_COMMIT)' \
           -X 'github.com/eddywiyatno/tm-agent/internal/buildinfo.BuildDate=$(BUILD_DATE)'

.PHONY: all build build-all test lint clean install validate help

all: test build

help:
	@echo "Targets:"
	@echo "  build       - Build native static binary for current OS/Arch"
	@echo "  build-all   - Cross-compile for Linux (amd64, arm64) and Windows (amd64)"
	@echo "  test        - Run unit and component tests"
	@echo "  validate    - Validate repository layout and contracts"
	@echo "  clean       - Clean build artifacts"
	@echo "  install     - Install binary to ~/.local/bin"

build:
	@mkdir -p bin
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME) ./cmd/tm-agent

build-all:
	@mkdir -p bin/linux_amd64 bin/linux_arm64 bin/windows_amd64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/linux_amd64/$(BINARY_NAME) ./cmd/tm-agent
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/linux_arm64/$(BINARY_NAME) ./cmd/tm-agent
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/windows_amd64/$(BINARY_NAME).exe ./cmd/tm-agent

test:
	go test -v -race=false ./...

validate:
	./scripts/validate.sh

install: build
	@mkdir -p $(HOME)/.local/bin
	cp bin/$(BINARY_NAME) $(HOME)/.local/bin/$(BINARY_NAME)
	@echo "Installed tm-agent to $(HOME)/.local/bin/$(BINARY_NAME)"

clean:
	rm -rf bin/ *.out *.tmp
