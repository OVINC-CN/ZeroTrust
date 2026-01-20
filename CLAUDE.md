# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ZeroTrust is a lightweight authentication gateway for CDN edge verification. It validates Django sessions stored in Redis and returns authorization decisions to CDN edge functions (Cloudflare Workers, Vercel Edge, etc.).

## Common Commands

```bash
# Build
make build                    # Build binary to bin/zerotrust
go build -o bin/zerotrust ./cmd/zerotrust

# Run
make run                      # Build and run with configs/config.yaml
./bin/zerotrust -config configs/config.yaml

# Test
make test                     # Run all tests
go test -v ./...              # Verbose test output
go test -v -race ./...        # Test with race detector

# Lint
go vet ./...                  # Go vet
golangci-lint run             # Full linting (used in CI)
staticcheck ./...             # Static analysis

# Dependencies
make deps                     # go mod tidy && go mod download
```

## Architecture

```
cmd/zerotrust/main.go    - Entry point, server setup, graceful shutdown
internal/
  config/config.go       - YAML config loading, global config singleton
  handler/verify.go      - POST /verify endpoint, session validation logic
  session/parser.go      - Django pickle session parser (uses gopickle)
  store/redis.go         - Redis client with OpenTelemetry tracing
  otel/otel.go           - OpenTelemetry initialization (trace only, no metrics)
  otel/middleware.go     - HTTP middleware for tracing
  log/log.go             - Logrus wrapper with trace hook, fixed-field logging
```

## Key Design Decisions

- **Logging**: Uses logrus with fixed fields (`module`, `action`, `error`, `data`) to keep ES mappings stable. Dynamic data goes in `data` object. TraceHook auto-injects `trace_id`/`span_id` from context.
- **OpenTelemetry**: Trace-only (no metrics). Resource attributes configurable via `otel.resource` in config.
- **Session Parsing**: Parses Python pickle format using gopickle. Extracts `_auth_user_id`, `_auth_user_backend`, `_auth_user_hash` from Django session dict.
- **Redis**: Uses go-redis with redisotel for automatic trace propagation.

## Configuration

Config file: `configs/config.yaml` (copy from `config.example.yaml`)

Key settings:
- `redis.session_key_format`: Redis key pattern with `{session_id}` placeholder
- `auth.unauthorized_action`: `"redirect"` or `"deny"`
- `otel.resource`: Configure service_name, service_version, environment, custom attributes

## API Endpoints

- `POST /verify` - Validate session, returns `{allowed, user_id, action, login_url, message}`
- `GET /health` - Health check, returns `ok`
