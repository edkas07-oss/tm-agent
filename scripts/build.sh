#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "${SCRIPT_DIR}")"

cd "${PROJECT_ROOT}"

VERSION="$(cat VERSION 2>/dev/null || echo "0.1.0")"
GIT_COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")"
BUILD_DATE="$(date -u +'%Y-%m-%dT%H:%M:%SZ')"

LDFLAGS="-s -w -X 'github.com/eddywiyatno/tm-agent/internal/buildinfo.Version=${VERSION}' \
            -X 'github.com/eddywiyatno/tm-agent/internal/buildinfo.GitCommit=${GIT_COMMIT}' \
            -X 'github.com/eddywiyatno/tm-agent/internal/buildinfo.BuildDate=${BUILD_DATE}'"

export PATH="${HOME}/.local/bin:${HOME}/.local/go/bin:${PATH}"

echo "Building native binary..."
mkdir -p bin
CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o bin/tm-agent ./cmd/tm-agent

echo "Cross-compiling for Linux (amd64, arm64) and Windows (amd64)..."
mkdir -p bin/linux_amd64 bin/linux_arm64 bin/windows_amd64

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o bin/linux_amd64/tm-agent ./cmd/tm-agent
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o bin/linux_arm64/tm-agent ./cmd/tm-agent
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o bin/windows_amd64/tm-agent.exe ./cmd/tm-agent

echo "4. Generating SHA-256 checksums manifest..."
(
    cd "${PROJECT_ROOT}/bin"
    sha256sum linux_amd64/tm-agent linux_arm64/tm-agent windows_amd64/tm-agent.exe tm-agent > checksums.txt
)

echo "Build complete. Artifacts:"
ls -lh bin/linux_amd64/tm-agent bin/linux_arm64/tm-agent bin/windows_amd64/tm-agent.exe bin/tm-agent bin/checksums.txt

