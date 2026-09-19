# 📦 Installation & Deployment Guide — `tm-agent`

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8.svg)](https://go.dev)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows%20Server-blue.svg)](README.md)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Security](https://img.shields.io/badge/Security-Rootless_0700_Spool-brightgreen.svg)](CONFIG)

This guide provides comprehensive instructions for building, installing, configuring, and verifying the **`tm-agent`** (*Unified Cross-Platform Event Collector Daemon*) across Linux systems (*systemd user services*) and Windows Server hosts (*Windows background tasks/services*).

---

## 📑 Table of Contents

- [1. System & Container Engine Prerequisites](#1-system--container-engine-prerequisites)
- [2. Quick Start Compilation (Native Binary)](#2-quick-start-compilation-native-binary)
- [3. Local PATH Installation (`~/.local/bin`)](#3-local-path-installation-localbin)
- [4. Multi-OS Cross-Compilation Matrix](#4-multi-os-cross-compilation-matrix)
- [5. Systemd User Daemon Setup (Linux)](#5-systemd-user-daemon-setup-linux)
- [6. Windows Background Task & Service Setup (Windows Server)](#6-windows-background-task--service-setup-windows-server)
- [7. Spool Directory & Permissions Setup](#7-spool-directory--permissions-setup)
- [8. Post-Installation Verification & Health Checks](#8-post-installation-verification--health-checks)

---

## 1. System & Container Engine Prerequisites

### Supported Operating Systems
* **Linux:**
  * ✅ **Ubuntu Linux (20.04 / 22.04 / 24.04 LTS), Debian (11 / 12), elementary OS:** 100% Tested & Verified via `systemd --user`.
  * ✅ **Amazon Linux 2023 (AL2023), RHEL / Rocky Linux (8 / 9):** 100% Tested & Verified.
  * ✅ **Linux ARM64 / AWS Graviton:** 100% Compiled & Verified.
* **Windows:**
  * ✅ **Windows Server 2019 / 2022 / 2025 Datacenter:** 100% Tested & Verified via Docker Named Pipe (`\\.\pipe\docker_engine`).
  * 🔹 **Windows 10 / 11 Desktop (PowerShell):** Supported via Docker Desktop or Podman Machine.

### Container Runtimes Supported
* **Podman:** Version 4.0+ (Rootless streaming via `/run/user/<uid>/podman/podman.sock`).
* **Docker Engine / Mirantis Container Runtime:** Version 24.0+ (Linux Unix socket or Windows named pipe).

### Build Tools Required (Build from Source Only)
* **Go Compiler:** Version 1.23+ with `CGO_ENABLED=0` capability.
* **GNU Make & Coreutils:** For automated Makefile targets.

---

## 2. Quick Start Compilation (Native Binary)

To build the static daemon binary for the current host architecture:

```bash
# 1. Clone repository
git clone git@github.com:edkas07-oss/tm-agent.git
cd tm-agent

# 2. Build native static binary
make build
```

The compiled binary will be placed at `bin/tm-agent` (Linux) or `bin/tm-agent.exe` (Windows).

> [!NOTE]
> `tm-agent` is compiled with `CGO_ENABLED=0` and stripped symbols (`-s -w`), producing a single, self-contained static binary with zero external shared library dependencies.

---

## 3. Local PATH Installation (`~/.local/bin`)

To install `tm-agent` directly into your user's executable path:

```bash
# Install to ~/.local/bin/tm-agent
make install

# Verify PATH resolution
export PATH="${HOME}/.local/bin:${PATH}"
tm-agent --version
```

---

## 4. Multi-OS Cross-Compilation Matrix

To generate release-grade static binaries for all supported enterprise architectures simultaneously:

```bash
make build-all
```

### Generated Artifact Matrix:

| Output Binary | Target Architecture | Format | Size | Target Environment |
| :--- | :--- | :--- | :---: | :--- |
| `bin/linux_amd64/tm-agent` | `linux/amd64` | ELF 64-bit Static | ~6.1 MB | Linux Server (x86_64), systemd daemon |
| `bin/linux_arm64/tm-agent` | `linux/arm64` | ELF 64-bit Static | ~5.6 MB | AWS Graviton, ARM64 Edge Nodes |
| `bin/windows_amd64/tm-agent.exe` | `windows/amd64` | PE32+ Executable | ~6.4 MB | Windows Server 2019/2022/2025 |
| `bin/checksums.txt` | All | SHA-256 Manifest | - | Cryptographic Integrity Checksums |

---

## 5. Systemd User Daemon Setup (Linux)

To run `tm-agent` as a managed, auto-restarting background daemon on Linux without root privileges:

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

> [!TIP]
> **Enable Systemd Lingering for Rootless Daemons:**
> To ensure `tm-agent.service` starts on boot and continues running after user logout, enable lingering on your user account:
> ```bash
> loginctl enable-linger $(whoami)
> ```

---

## 6. Windows Background Task & Service Setup (Windows Server)

### Interactive Execution (PowerShell):
```powershell
.\bin\windows_amd64\tm-agent.exe --spool-dir C:\tm-home\spool
```

### Automated Background Startup Task Registration:
```powershell
# Create background scheduled startup task running under SYSTEM
$Action = New-ScheduledTaskAction -Execute "C:\tm-home\bin\tm-agent.exe" -Argument "--spool-dir C:\tm-home\spool --engine docker"
$Trigger = New-ScheduledTaskTrigger -AtStartup
Register-ScheduledTask -TaskName "TomcatMonitoringAgent" -Action $Action -Trigger $Trigger -User "SYSTEM"
Start-ScheduledTask -TaskName "TomcatMonitoringAgent"
```

---

## 7. Spool Directory & Permissions Setup

`tm-agent` writes evidence records into a secure persistent host spool directory:

```text
Spool Directory Structure (0700):
/opt/tm-home/spool/ (Linux)  or  C:\tm-home\spool\ (Windows)
├── 1789347140608638731_container_state.json  (0600 - Verified Schema)
├── 1789347140608775839_runtime_oom.json       (0600 - Verified Schema)
└── .tmp files                                 (Auto-pruned after 60 min)
```

Ensure permissions are securely initialized:
```bash
# Linux
mkdir -p /opt/tm-home/spool
chmod 700 /opt/tm-home/spool

# Windows Server (PowerShell)
New-Item -ItemType Directory -Force -Path C:\tm-home\spool
```

---

## 8. Post-Installation Verification & Health Checks

Verify agent execution and socket streaming:

```bash
# 1. Execute one-shot test snapshot
tm-agent --run-once --spool-dir /tmp/test-spool --target tomcat-jmx-exporter

# 2. Verify generated JSON evidence records
ls -la /tmp/test-spool/
cat /tmp/test-spool/*_container_state.json

# 3. Validate repository layout and platform governance rules
make validate
```
