# a2a-playground

A local CLI that serves a Vue UI and proxies to an A2A (Agent-to-Agent) agent. Use it to test and interact with A2A agents in your browser. Supports both **gRPC** and **JSON-RPC** transports.

## Architecture

```
┌─────────┐     ┌─────────────────────────────────────┐     ┌────────────────────────────┐
│ Browser │────▶│ BFF (static SPA + Connect proxy)    │────▶│ gRPC or JSON-RPC A2A agent  │
│         │     │ serves Vue app, proxies A2A RPC      │     │ (e.g. localhost:8080)      │
└─────────┘     └─────────────────────────────────────┘     └────────────────────────────┘
```

The BFF (Backend-for-Frontend) serves the static Vue SPA and proxies Connect-RPC requests to your A2A agent. You can use **gRPC** (default) or **JSON-RPC** depending on what your agent exposes.

## Prerequisites

- **Go** 1.25+
- **Node.js** 20+
- **pnpm**
- **buf** CLI (for proto generation)

## Installation

### With Go

Install the binary to `$GOPATH/bin` (or `$HOME/go/bin`):

```bash
go install github.com/alis-exchange/a2a-playground/cmd/a2a-playground@latest
```

Or a specific version:

```bash
go install github.com/alis-exchange/a2a-playground/cmd/a2a-playground@v1.0.0
```

Run without installing (fetches and builds on first use):

```bash
# gRPC
go run github.com/alis-exchange/a2a-playground/cmd/a2a-playground@latest --agent-url=localhost:8080

# JSON-RPC
go run github.com/alis-exchange/a2a-playground/cmd/a2a-playground@latest --agent-url=http://localhost:8080/jsonrpc --jsonrpc
```

> **Note:** `go run` and `go install` with a version (e.g. `@v1.0.0`) only work for releases where the embedded frontend was committed. Use the [install script](#from-release) or [build from source](#from-source) if a version fails with "no matching files found".

### From release

Install the latest version:

```bash
curl -fsSL https://raw.githubusercontent.com/alis-exchange/a2a-playground/main/install.sh | bash
```

Or install a specific version:

```bash
./install.sh -v v1.0.0
```

### From source

```bash
make build
```

This generates protobuf code, builds the frontend, embeds it, and compiles the `a2a-playground` binary.

## Usage

Run the playground with an A2A agent. The browser opens automatically at `http://localhost:3000` (or the port you specify).

### gRPC (default)

```bash
# gRPC: agent-url is host:port
a2a-playground --agent-url=localhost:8080
```

### JSON-RPC

```bash
# JSON-RPC: agent-url is full HTTP URL including path
a2a-playground --agent-url=http://localhost:8080/jsonrpc --jsonrpc
```

### Flags

| Flag          | Default                   | Description                                                                                          |
| ------------- | ------------------------- | ---------------------------------------------------------------------------------------------------- |
| `--agent-url` | `localhost:8080`          | Agent endpoint. For gRPC: `host:port`. For JSON-RPC: full URL (e.g. `http://localhost:8080/jsonrpc`) |
| `--grpc`      | _(when no protocol flag)_ | Use gRPC transport (default)                                                                         |
| `--jsonrpc`   | —                         | Use JSON-RPC over HTTP transport                                                                     |
| `--port`      | `3000`                    | HTTP port for the BFF                                                                                |
| `--no-open`   | `false`                   | Do not open the browser on start                                                                     |
| `--dev`       | `false`                   | Serve from `app/dist` on disk instead of embedded files                                              |

### Custom headers

Configure authentication and custom headers in the playground UI (key icon in the toolbar). Headers such as `Authorization`, `X-API-Key`, and `X-Tenant-ID` are persisted and forwarded to the agent on every request.

## Development

### Run in dev mode

Uses `app/dist` on disk instead of embedded files (no rebuild needed after frontend changes):

```bash
make run
```

This runs with `--dev --agent-url=localhost:8080`.

### Generate protobuf only

```bash
make generate
```

Runs `buf generate` in `packages/a2a` (generates TS into `app/packages/a2a/protobuf` and Go into `gen/go`).

### Frontend development

```bash
cd app && pnpm dev
```

Build frontend for production:

```bash
cd app && pnpm build
```

## Releasing

The binary embeds `app/dist` at build time. For `go run` and `go install` with a version to work, the built frontend must be committed **before** tagging:

```bash
make generate build-frontend
git add app/dist
git commit -m "chore: embed frontend for v1.0.0"
git tag v1.0.0
git push origin main --tags
```

The release workflow then builds the cross-platform binaries and attaches them to the GitHub release.

## Project structure

| Directory             | Description                                                  |
| --------------------- | ------------------------------------------------------------ |
| `cmd/a2a-playground/` | CLI entrypoint                                               |
| `internal/bff/`       | BFF server, static serving, Connect proxy (gRPC or JSON-RPC) |
| `packages/a2a/`       | Proto definitions and buf config (A2A canonical)             |
| `app/`                | Vue 3 + Vuetify SPA                                          |
| `gen/go/`             | Generated Connect handlers (uses `a2a-go/a2apb`)             |

## Contributing

1. Fork and open a PR
2. Use [conventional commits](https://www.conventionalcommits.org/)
3. Run tests: `go test ./...`

## License

See [LICENSE](LICENSE).
