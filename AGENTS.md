# AGENTS.md — Developer & AI Agent Guidelines for `tm-agent`

## 🎯 Repository Purpose
`tm-agent` is a unified, lightweight, cross-platform **Event Collector Daemon** written in Go. It consumes real-time container lifecycle streaming events directly from the **Container Engine Socket API** (Podman / Docker) and writes schema-compliant JSON evidence records atomically into a persistent host spool directory according to the `event-record-v1.schema.json` contract.

## 🏛️ Architecture Rules & Non-Negotiables
1. **Zero External Runtime Dependency:** `tm-agent` MUST be compiled as a single static binary (`CGO_ENABLED=0`) without runtime dependencies on external shells or utilities.
2. **Direct Socket Streaming:** Consumes `GET /events` (Docker) or `GET /v4.0.0/libpod/events` (Podman) directly via Unix Domain Sockets, Windows Named Pipes (`\\.\pipe\docker_engine`), or TCP mTLS.
3. **100% Identical Data Contract:** Event snapshots must strictly follow the canonical `event-record-v1.schema.json` with atomic writing (`.tmp` $\rightarrow$ `.json`, mode `0600`) in a `0700` spool directory.
4. **Autonomous FIFO Spool Retention:** Enforce 24-hour expiration pruning, 1,000-file FIFO quota limits, and 60-minute orphan `.tmp` cleanup.
5. **Cross-Platform Parity:** Runs natively as a systemd user daemon on Linux and as a background service/process on Windows Server.
6. **Cross-Compilation Matrix:** Supports compilation for Linux (`amd64`, `arm64`) and Windows (`amd64` producing `tm-agent.exe`).

## 🛠️ Build & Validation Commands
- **Build Native:** `make build`
- **Cross-Compile:** `make build-all`
- **Unit Test Suite:** `make test`
- **Static Validation:** `make validate`
