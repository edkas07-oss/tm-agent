# AGENTS.md — Agent Governance & Instructions for tm-agent

## Overview
`tm-agent` adalah agen background pengumpul event kontainer (*Unified Cross-Platform Event Collector Daemon*) berbasis bahasa Go yang mengonsumsi *real-time streaming events* langsung dari **Container Engine Socket API** (Podman / Docker) dan menulis berkas bukti insiden (*evidence spool*) secara atomik sesuai kontrak kanonikal `event-record-v1.schema.json`.

## Architecture Rules & Constraints
1. **Zero External Runtime Dependency:** `tm-agent` dikompilasi sebagai *Single Static Binary* (`CGO_ENABLED=0`) tanpa ketergantungan runtime (tanpa Bash, jq, atau Linux coreutils).
2. **Direct Socket Streaming:** Mengonsumsi stream `GET /events` atau `GET /v4.0.0/libpod/events` langsung via Unix Domain Socket, Windows Named Pipe (`\\.\pipe\docker_engine`), atau TCP mTLS.
3. **Kontrak Data 100% Identik (*Zero Breaking Change*):** Snapshot event wajib mematuhi skema kanonikal `event-record-v1.schema.json` dengan penulisan atomik (`.tmp` -> `.json` izin `0600`) pada direktori spool `0700`.
4. **FIFO Spool Retention & Quota Guard:** Membatasi usia berkas spool (maksimal 24 jam) dan batas kuota berkas (FIFO pruning jika > 1000 berkas), serta membersihkan berkas `.tmp` terlantar (> 60 menit).
5. **Cross-Platform Runner:** Mendukung eksekusi sebagai daemon foreground/systemd di Linux dan Windows Service / background process di Windows.
6. **Cross-Compilation Matrix:** Mendukung kompilasi silang ke Linux (`amd64`, `arm64`) dan Windows (`amd64` menghasilkan `tm-agent.exe`).
