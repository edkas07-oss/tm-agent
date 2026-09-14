#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "${SCRIPT_DIR}")"

cd "${PROJECT_ROOT}"

echo "Running repository contract validation..."

# 1. Check required contract files
required_files=("CONFIG" "PROJECT" "VERSION" "AGENTS.md" "README.md" "Makefile" "Jenkinsfile")
for f in "${required_files[@]}"; do
    if [[ ! -f "${f}" ]]; then
        echo "ERROR: Missing required contract file: ${f}" >&2
        exit 1
    fi
done

# 2. Check for sensitive files
sensitive_matches="$(find . -type f \( -name "*.pem" -o -name "*.key" -o -name "*.p12" -o -name "*.pfx" -o -name ".env*" -o -name "id_rsa" \) | grep -v "/.git/" || true)"
if [[ -n "${sensitive_matches}" ]]; then
    echo "ERROR: Forbidden sensitive material detected:" >&2
    echo "${sensitive_matches}" >&2
    exit 1
fi

# 3. Check Go formatting
unformatted="$(gofmt -l . || true)"
if [[ -n "${unformatted}" ]]; then
    echo "WARNING: Unformatted Go files detected:" >&2
    echo "${unformatted}" >&2
    gofmt -w .
fi

echo "Repository validation PASSED."
