# a2a-playground

A local CLI that serves a Vue UI and proxies to an A2A (Agent-to-Agent) agent. Use it to test and interact with A2A agents in your browser. Supports both **gRPC** and **JSON-RPC** transports.

## Quick start

With an A2A agent running (e.g. on port 8080), install and run:

```bash
go install github.com/alis-exchange/a2a-playground/cmd/a2a-playground@latest
a2a-playground --agent-url=localhost:8080
```

The browser opens at `http://localhost:3000`. Start chatting with your agent.

## Installation

### From release

No dependencies. Installs the binary for your platform:

```bash
curl -fsSL https://raw.githubusercontent.com/alis-exchange/a2a-playground/main/install.sh | bash
```

Or a specific version:

```bash
./install.sh -v v1.0.0
```

### With Go

Requires **Go 1.25+**:

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

### From source

Requires **Go 1.25+**, **Node.js 20+**, **pnpm**, and **buf** CLI:

```bash
make build
```

This generates protobuf code, builds the frontend, embeds it, and compiles the `a2a-playground` binary.

## Usage

### gRPC (default)

```bash
a2a-playground --agent-url=localhost:8080
```

### JSON-RPC

```bash
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

## Architecture

```
┌─────────┐     ┌─────────────────────────────────────┐     ┌────────────────────────────┐
│ Browser │────▶│ BFF (static SPA + Connect proxy)    │────▶│ gRPC or JSON-RPC A2A agent  │
│         │     │ serves Vue app, proxies A2A RPC      │     │ (e.g. localhost:8080)      │
└─────────┘     └─────────────────────────────────────┘     └────────────────────────────┘
```

The BFF (Backend-for-Frontend) serves the static Vue SPA and proxies Connect-RPC requests to your A2A agent. You can use **gRPC** (default) or **JSON-RPC** depending on what your agent exposes.

### Message flow (Frontend → BFF → Agent)

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│ Frontend (Vue)                                                                      │
│  messages.ts → createA2AClient().sendStreamingMessage(req)                          │
│  a2aClient.ts: agentHeadersInterceptor sets X-A2A-Agent-Headers from Pinia store    │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        │ HTTP POST (Connect) to /a2a.v1.A2AService/*
                                        │ Header: X-A2A-Agent-Headers: {"Authorization":"Bearer x"}
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│ BFF                                                                                  │
│  server.go: ExtractAgentHeaders(r) → WithAgentHeaders(ctx) → proxy handler          │
│  Protocol switch: cfg.Protocol ? NewJSONRPCProxy : NewGrpcProxy                     │
└─────────────────────────────────────────────────────────────────────────────────────┘
                    │                                    │
        ┌───────────┴───────────┐            ┌────────────┴────────────┐
        │ grpcProxy              │            │ jsonrpcProxy             │
        │ metadata.Append(...)   │            │ agentHeadersInterceptor  │
        │ → gRPC agent           │            │ req.Meta.Append(...)     │
        │ localhost:8080         │            │ → HTTP JSON-RPC agent    │
        └───────────────────────┘            │ http://.../jsonrpc       │
                                              └─────────────────────────┘
```

### How it works

**1. Frontend**

- User input flows through `PlaygroundView` into the messages store (`app/src/pages/playground/store/messages.ts`), which builds a `SendMessageRequest` and calls `createA2AClient().sendStreamingMessage(request)`.
- The Connect client (`app/src/clients/a2aClient.ts`) uses same-origin requests, so all A2A calls go to the BFF (e.g. `http://localhost:3000`).
- Before every request, the `agentHeadersInterceptor` reads the Pinia store (`app/src/store/agentHeaders.ts`), serializes configured headers to JSON, and sets the `X-A2A-Agent-Headers` HTTP header. Headers are configured in the UI (key icon) and can be persisted to localStorage.

**2. BFF routing and headers**

- The BFF (`internal/bff/server.go`) mounts two handler groups: the A2A Connect proxy at `/a2a.v1.A2AService/` and the static SPA at `/`.
- All A2A requests pass through middleware that reads `X-A2A-Agent-Headers`, parses it as `map[string]string`, and puts it into the request context (`internal/bff/headers.go`). The chosen proxy can then read these headers.

**3. Protocol switching**

- The CLI (`cmd/a2a-playground/main.go`) sets `--jsonrpc` or defaults to gRPC. That value is passed into `ServerConfig.Protocol`.
- `NewServer` checks `cfg.Protocol`: if JSON-RPC, it creates `NewJSONRPCProxy(cfg.AgentURL)`; otherwise `NewGrpcProxy(cfg.AgentURL)`. Both proxies implement `A2AServiceHandler`, so routing stays the same; only the outbound transport differs.

**4. gRPC proxy**

- `internal/bff/proxy.go`: The gRPC proxy uses `go.alis.build/client` to connect to the agent. Before each call, `withAgentHeaders(ctx)` reads headers from context and adds them to `metadata.NewOutgoingContext`. Those metadata entries become gRPC headers on the agent call. The proxy forwards Connect requests as native gRPC calls.

**5. JSON-RPC proxy**

- `internal/bff/jsonrpc_proxy.go`: The JSON-RPC proxy uses `a2aclient` from `a2a-go` with `WithJSONRPCTransport`. It converts Connect/protobuf requests to JSON-RPC using `pbconv.FromProto*` and `pbconv.ToProto*`. The proxy is built with `WithInterceptors(&agentHeadersInterceptor{})`, which reads headers from context and appends them to `req.Meta`; the a2a-go client sends `Meta` as HTTP headers to the agent's JSON-RPC endpoint. `ListTasks` is not supported over JSON-RPC per the A2A spec and returns `CodeUnimplemented`.

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
| `packages/a2a/`       | Proto definitions and buf config (A2A canonical)              |
| `app/`                | Vue 3 + Vuetify SPA                                          |
| `gen/go/`             | Generated Connect handlers (uses `a2a-go/a2apb`)              |

## Contributing

1. Fork and open a PR
2. Use [conventional commits](https://www.conventionalcommits.org/)
3. Run tests: `go test ./...`

## License

See [LICENSE](LICENSE).
