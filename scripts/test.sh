#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "${SCRIPT_DIR}")"

cd "${PROJECT_ROOT}"

export PATH="${HOME}/.local/bin:${HOME}/.local/go/bin:${PATH}"

echo "Running Go unit and component tests..."
go test -v -race=false ./...

echo "All tests passed successfully."
