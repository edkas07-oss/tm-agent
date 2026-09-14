# tm-agent — Unified Cross-Platform Event Collector Daemon

`tm-agent` adalah agen pengumpul event insiden (*Event Collector Daemon*) mandiri berbasis Go yang beroperasi secara *real-time* dengan mengonsumsi *event stream* langsung dari **Container Engine Socket API** (Podman / Docker) dan menuliskan bukti kejadian ke direktori *spool* persisten secara atomik.

Kakas ini merealisasikan **TASK-TM-028 (TN-013)** berdasarkan keputusan arsitektur [TM-ADR-0027](file:///home/eddywiyatno/git/devops-handbook/docs/adr/tomcat-monitoring/adr-records/TM-ADR-0027.md) dan [TM-ADR-0008](file:///home/eddywiyatno/git/devops-handbook/docs/adr/tomcat-monitoring/adr-records/TM-ADR-0008.md).

---

## 🚀 Fitur Utama

- **Single Static Binary (`CGO_ENABLED=0`):** Beroperasi mandiri tanpa dependensi runtime eksternal (tanpa Python, Bash, atau Linux coreutils).
- **Direct Socket Event Streaming:** Berlangganan langsung ke endpoint `GET /events` (Docker) atau `GET /v4.0.0/libpod/events` (Podman) via Unix Domain Socket, Windows Named Pipe (`\\.\pipe\docker_engine`), atau TCP.
- **Kepatuhan Kontrak Skema Kanonikal:** Memformat rekam bukti insiden (`container_state`, `runtime_oom`, `collector_status`) 100% identik dengan `event-record-v1.schema.json`.
- **Penulisan Atomik & Isolasi Izin:** Menulis ke berkas `.tmp` dengan izin `0600` sebelum melakukan `rename` atomik ke `.json` pada direktori spool berizin `0700`.
- **FIFO Spool Retention & Quota Guard:** Pemangkasan otomatis berkas `.json` (> 24 jam), pembersihan berkas `.tmp` terlantar (> 60 menit), dan penegakan kuota FIFO jika jumlah berkas > 1000.
- **Cross-Platform Execution:** Mendukung eksekusi sebagai daemon foreground/systemd di Linux dan Windows Service / console runner di Windows.

---

## 📦 Penggunaan CLI

```bash
# Menjalankan agen secara foreground (listening to event stream)
tm-agent

# Mengambil one-shot snapshot dan keluar
tm-agent --run-once

# Menentukan target container dan spool directory secara eksplisit
tm-agent --target tomcat-jmx-exporter --target-id lab/tomcat-01/default --spool-dir /var/lib/monitoring/spool

# Menentukan engine dan socket path kustom
tm-agent --engine podman --socket /run/user/1000/podman/podman.sock

# Menampilkan informasi versi biner
tm-agent --version
```

---

## 🛠️ Kompilasi & Pengujian

```bash
# Kompilasi native binary
make build

# Kompilasi silang (Linux amd64, Linux arm64, Windows amd64)
make build-all

# Menjalankan seluruh pengujian unit
make test

# Menjalankan validasi tata kelola repositori
make validate
```
