# OCULTAR | Release & Distribution Guide

This guide explains how to build and distribute OCULTAR binaries or Docker images to clients.

---

## 1. Build the Release

The `build_release.sh` script compiles and packages the distribution in one step:

```bash
./build_release.sh
```

Output in `dist/`: versioned archives for Linux, macOS, and Windows, built by the GitHub Actions `release.yml` workflow and published as GitHub Releases.

> [!NOTE]
> OCULTAR runs all PII detection locally — local SLM-based NER, on-premise vault, zero internet required at runtime. The only external dependency is the one-time model download from HuggingFace on first run.
> These archives are generated automatically by CI and are **not tracked in version control** to prevent "dirty" repository loops. They are intended for distribution via GitHub Releases.

---

## 2. Distribution Options

### Option A: Docker (Recommended for All Clients)

Docker standardizes the environment: clients don't need Go, Python, or any dependency.

**What to send:**
1. The distribution archive from the GitHub Releases page
2. `documentation/SETUP_GUIDE.md`

The client unzips, runs `setup.sh` (Linux/macOS) or `setup.ps1` (Windows), and opens `http://localhost:3030`.

**Alternative — Docker image tarball:**
```bash
docker build -t ocultar-api .
docker save ocultar-api > ocultar_v$(date +%Y%m%d).tar
```

The client loads it with:
```bash
docker load -i ocultar_v20260301.tar
docker compose up -d
```

*Works on Windows (WSL2), macOS, and Linux without modification.*

---

### Option B: Native Bare-Metal Binaries

For developers or security researchers who want to run the raw binary without Docker.

> **Note:** OCULTAR uses DuckDB (via CGO), so cross-compilation requires a C toolchain.

#### Linux (native, from a Linux machine)
```bash
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
  go build -o dist/community/ocultar ./dist/community
```

#### Windows (from Linux, using mingw-w64)
```bash
# Install cross-compiler first:
sudo apt-get install gcc-mingw-w64

# Build:
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
  CC=x86_64-w64-mingw32-gcc \
  go build -o dist/community/ocultar.exe ./dist/community
```

#### macOS (Silicon / Intel)
Cross-compiling for macOS requires Apple's proprietary SDKs. Three practical options:

1. **Build on an actual Mac** — easiest:
   ```bash
   CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
     go build -o dist/community/ocultar-macos-arm64 ./dist/community
   ```
2. **GitHub Actions** — set up a macOS runner in CI/CD to produce the binary automatically
3. **Recommend Docker** — Docker Desktop runs natively on Mac, sidestepping this entirely

---

## 3. Go Workspace Structure

The project uses a `go.work` file to manage a multi-module workspace:

```
go.work
├── apps/proxy       (github.com/ocultar-dev/ocultar/apps/proxy)   ← transparent proxy
├── apps/sombra      (github.com/ocultar-dev/ocultar/apps/sombra)  ← Sombra gateway
├── apps/slm-engine  (github.com/ocultar-dev/ocultar/apps/slm-engine)
├── services/refinery
├── services/vault
├── internal/pii
└── pkg/gateway
```

To build all modules from the root during development:
```bash
go build ./...
```

---

## 4. Proxy Distribution

The `docker-compose.proxy.yml` file is the deployment unit for the proxy mode. Include it alongside the binary for clients who want transparent LLM API interception:

```bash
# Proxy cluster startup:
docker compose -f docker-compose.proxy.yml up -d
```

See `docs/setup_guide.md` for the full proxy deployment guide.
