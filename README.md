<div align="center">
  <img src="docs/favicon.svg" alt="ZeroTrust Logo" width="64" height="64">
  <h1>ZeroTrust</h1>
  <p><strong>A lightweight Zero Trust authentication gateway for CDN edge verification</strong></p>

  [![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
  [![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
  [![Build](https://github.com/OVINC-CN/ZeroTrust/actions/workflows/build.yml/badge.svg)](https://github.com/OVINC-CN/ZeroTrust/actions/workflows/build.yml)
  [![Docker](https://github.com/OVINC-CN/ZeroTrust/actions/workflows/docker.yml/badge.svg)](https://github.com/OVINC-CN/ZeroTrust/actions/workflows/docker.yml)

  **English** | [中文](README_zh.md)
</div>

---

## Overview

ZeroTrust is a high-performance authentication verification service designed to work with CDN edge functions (Cloudflare Workers, Vercel Edge Functions, etc.). It validates user sessions against a Redis-backed Django session store, implementing Zero Trust security principles at the edge.

### Key Features

- **Edge-First Design** - Built for CDN edge function integration
- **Django Session Compatible** - Parses Django pickle-serialized sessions
- **Redis Backend** - Fast session lookup with configurable key format
- **OpenTelemetry Support** - Full observability with traces and metrics
- **Flexible Auth Actions** - Redirect to login or deny access
- **Lightweight** - Minimal dependencies, fast startup

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│  CDN Edge   │────▶│  ZeroTrust  │────▶│    Redis    │
│             │     │  Function   │     │   Gateway   │     │  (Session)  │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                           │                   │
                           │                   │ ✓ Valid Session
                           │                   ▼
                           │            ┌─────────────┐
                           └───────────▶│   Origin    │
                                        │   Server    │
                                        └─────────────┘
```

## Quick Start

### Prerequisites

- Go 1.24+
- Redis server
- Django application with Redis session backend

### Installation

```bash
# Clone the repository
git clone https://github.com/OVINC-CN/ZeroTrust.git
cd ZeroTrust

# Build the binary
make build

# Or using Go directly
go build -o bin/zerotrust ./cmd/zerotrust
```

### Configuration

Copy the example configuration and customize:

```bash
cp configs/config.example.yaml configs/config.yaml
```

```yaml
server:
  host: "0.0.0.0"
  port: 8080

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  session_key_format: ":1:django.contrib.sessions.cache{session_id}"

auth:
  unauthorized_action: "redirect"  # or "deny"
  login_url: "https://example.com/login"
  session_cookie_name: "sessionid"

otel:
  enabled: false
  endpoint: "localhost:4317"
  service_name: "zerotrust"
  insecure: true
```

### Running

```bash
# Run with default config
make run

# Or specify config path
./bin/zerotrust -config /path/to/config.yaml
```

### Docker

```bash
# Build image
docker build -t zerotrust:latest .

# Run container
docker run -p 8080:8080 -v $(pwd)/configs:/app/configs zerotrust:latest
```

Or pull from GitHub Container Registry:

```bash
docker pull ghcr.io/ovinc-cn/zerotrust:latest
```

## API Reference

### POST /verify

Verify a user session.

**Request:**

```json
{
  "session_id": "abc123xyz",
  "method": "GET",
  "path": "/api/users",
  "req_size": 1024,
  "params": {"page": "1"},
  "user_agent": "Mozilla/5.0...",
  "client_ip": "192.168.1.1",
  "host": "api.example.com",
  "referer": "https://example.com"
}
```

**Response (Authorized):**

```json
{
  "allowed": true,
  "user_id": "username",
  "message": "authorized"
}
```

**Response (Unauthorized):**

```json
{
  "allowed": false,
  "action": "redirect",
  "login_url": "https://example.com/login",
  "message": "unauthorized"
}
```

### GET /health

Health check endpoint.

**Response:** `200 OK` with body `ok`

## Edge Function Integration

### Cloudflare Workers Example

https://github.com/OVINC-CN/CFWorker/blob/main/remote-auth/index.js

## Observability

ZeroTrust supports OpenTelemetry for distributed tracing and metrics.

### Tracing

All HTTP requests are traced with the following attributes:
- `http.method`
- `http.url`
- `http.status_code`
- `http.user_agent`

## Development

```bash
# Install dependencies
make deps

# Build
make build

# Run tests
make test

# Clean build artifacts
make clean
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [go-redis](https://github.com/redis/go-redis) - Redis client for Go
- [gopickle](https://github.com/nlpodyssey/gopickle) - Python pickle decoder for Go
- [OpenTelemetry](https://opentelemetry.io/) - Observability framework