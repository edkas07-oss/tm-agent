# 📦 Installation & Build Guide — `tm-agent`

This document provides complete instructions for compiling, installing, and deploying the **`tm-agent`** (*Unified Cross-Platform Event Collector Daemon*) across Linux and Windows Server environments.

---

## 📑 Table of Contents

- [1. Prerequisites](#1-prerequisites)
- [2. Building from Source](#2-building-from-source)
- [3. Local Installation (`~/.local/bin`)](#3-local-installation-localbin)
- [4. Cross-Compilation Matrix](#4-cross-compilation-matrix)
- [5. Systemd User Daemon Setup (Linux)](#5-systemd-user-daemon-setup-linux)
- [6. Windows Service & Background Execution (Windows Server)](#6-windows-service--background-execution-windows-server)
- [7. Verification & Health Check](#7-verification--health-check)

---

## 1. Prerequisites

- **Go:** Version 1.23+ installed on build workstation.
- **Container Engine:** Podman 4.0+ or Docker Engine 24.0+ (running rootless or system socket).
- **Target OS:** Linux (`amd64` / `arm64`) or Windows Server 2019/2022/2025 (`amd64`).

---

## 2. Building from Source

To compile the native static binary for your current operating system and architecture:

```bash
# Clone repository
git clone git@github.com:edkas07-oss/tm-agent.git
cd tm-agent

# Build native binary into bin/
make build
```

This compiles a pure static binary (`CGO_ENABLED=0`) into `bin/tm-agent` (or `bin/tm-agent.exe` on Windows).

---

## 3. Local Installation (`~/.local/bin`)

To install `tm-agent` into your user's `$PATH`:

```bash
make install
```

This copies the binary to `~/.local/bin/tm-agent`. Ensure `~/.local/bin` is in your `$PATH`:
```bash
export PATH="${HOME}/.local/bin:${PATH}"
tm-agent --version
```

---

## 4. Cross-Compilation Matrix

To build the full matrix of production binaries for all supported platforms:

```bash
make build-all
```

Outputs generated in `bin/`:
- `bin/linux_amd64/tm-agent` (Linux x86_64)
- `bin/linux_arm64/tm-agent` (Linux ARM64 / AWS Graviton)
- `bin/windows_amd64/tm-agent.exe` (Windows x86_64)
- `bin/checksums.txt` (SHA-256 integrity checksums)

---

## 5. Systemd User Daemon Setup (Linux)

To run `tm-agent` as a managed, auto-restarting background daemon on Linux:

```bash
# 1. Create systemd user unit directory
mkdir -p ~/.config/systemd/user

# 2. Copy service unit template
cp systemd/tm-agent.service ~/.config/systemd/user/

# 3. Reload systemd and enable service
systemctl --user daemon-reload
systemctl --user enable --now tm-agent.service

# 4. Check service status and live stream logs
systemctl --user status tm-agent.service
journalctl --user -u tm-agent.service -f
```

---

## 6. Windows Service & Background Execution (Windows Server)

### Interactive Execution (PowerShell):
```powershell
.\bin\windows_amd64\tm-agent.exe --spool-dir C:\tm-home\spool
```

### Windows Background Task Registration:
```powershell
# Create background scheduled startup task
$Action = New-ScheduledTaskAction -Execute "C:\tm-home\bin\tm-agent.exe" -Argument "--spool-dir C:\tm-home\spool"
$Trigger = New-ScheduledTaskTrigger -AtStartup
Register-ScheduledTask -TaskName "TomcatMonitoringAgent" -Action $Action -Trigger $Trigger -User "SYSTEM"
```

---

## 7. Verification & Health Check

Verify agent connection to the engine socket and one-shot snapshot generation:

```bash
# Execute one-shot test snapshot
tm-agent --run-once --spool-dir /tmp/test-spool

# Verify generated JSON evidence records
ls -la /tmp/test-spool/
cat /tmp/test-spool/*_container_state.json
```
