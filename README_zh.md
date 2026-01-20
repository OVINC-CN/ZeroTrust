<div align="center">
  <img src="docs/favicon.svg" alt="ZeroTrust Logo" width="64" height="64">
  <h1>ZeroTrust</h1>
  <p><strong>轻量级零信任身份验证网关，专为 CDN 边缘验证设计</strong></p>

  [![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
  [![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
  [![Build](https://github.com/OVINC/ZeroTrust/actions/workflows/build.yml/badge.svg)](https://github.com/OVINC/ZeroTrust/actions/workflows/build.yml)
  [![Docker](https://github.com/OVINC/ZeroTrust/actions/workflows/docker.yml/badge.svg)](https://github.com/OVINC/ZeroTrust/actions/workflows/docker.yml)

  [English](README.md) | **中文**
</div>

---

## 概述

ZeroTrust 是一个高性能的身份验证服务，专为 CDN 边缘函数（Cloudflare Workers、Vercel Edge Functions 等）设计。它通过 Redis 后端验证 Django Session，在边缘实现零信任安全原则。

### 核心特性

- **边缘优先设计** - 专为 CDN 边缘函数集成打造
- **Django Session 兼容** - 解析 Django pickle 序列化的 Session
- **Redis 后端** - 快速 Session 查询，支持可配置的 Key 格式
- **OpenTelemetry 支持** - 完整的可观测性，包含链路追踪和指标
- **灵活的认证策略** - 支持跳转登录或拒绝访问
- **轻量级** - 最小依赖，快速启动

## 架构

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│    客户端    │────▶│  CDN 边缘   │────▶│  ZeroTrust  │────▶│    Redis    │
│             │     │    函数     │     │     网关     │     │  (Session)  │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                           │                   │
                           │                   │ ✓ 有效 Session
                           │                   ▼
                           │            ┌─────────────┐
                           └───────────▶│   源站服务   │
                                        │             │
                                        └─────────────┘
```

## 快速开始

### 环境要求

- Go 1.24+
- Redis 服务器
- 使用 Redis Session 后端的 Django 应用

### 安装

```bash
# 克隆仓库
git clone https://github.com/OVINC-CN/ZeroTrust.git
cd ZeroTrust

# 构建二进制文件
make build

# 或直接使用 Go
go build -o bin/zerotrust ./cmd/zerotrust
```

### 配置

复制示例配置文件并自定义：

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
  unauthorized_action: "redirect"  # 或 "deny"
  login_url: "https://example.com/login"
  session_cookie_name: "sessionid"

otel:
  enabled: false
  endpoint: "localhost:4317"
  service_name: "zerotrust"
  insecure: true
```

### 运行

```bash
# 使用默认配置运行
make run

# 或指定配置文件路径
./bin/zerotrust -config /path/to/config.yaml
```

### Docker

```bash
# 构建镜像
docker build -t zerotrust:latest .

# 运行容器
docker run -p 8080:8080 -v $(pwd)/configs:/app/configs zerotrust:latest
```

或从 GitHub Container Registry 拉取：

```bash
docker pull ghcr.io/ovinc-cn/zerotrust:latest
```

## API 参考

### POST /verify

验证用户 Session。

**请求：**

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

**响应（已授权）：**

```json
{
  "allowed": true,
  "user_id": "username",
  "message": "authorized"
}
```

**响应（未授权）：**

```json
{
  "allowed": false,
  "action": "redirect",
  "login_url": "https://example.com/login",
  "message": "unauthorized"
}
```

### GET /health

健康检查端点。

**响应：** `200 OK`，响应体为 `ok`

## 边缘函数集成

### 腾讯云 EdgeOne 边缘函数示例

https://github.com/OVINC-CN/CFWorker/blob/main/remote-auth/index.js

## 可观测性

ZeroTrust 支持 OpenTelemetry 分布式追踪和指标收集。

### 链路追踪

所有 HTTP 请求都会被追踪，包含以下属性：
- `http.method`
- `http.url`
- `http.status_code`
- `http.user_agent`

## 开发

```bash
# 安装依赖
make deps

# 构建
make build

# 运行测试
make test

# 清理构建产物
make clean
```

## 贡献

欢迎贡献代码！请随时提交 Pull Request。

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 许可证

本项目基于 MIT 许可证开源 - 详见 [LICENSE](LICENSE) 文件。

## 致谢

- [go-redis](https://github.com/redis/go-redis) - Go 语言 Redis 客户端
- [gopickle](https://github.com/nlpodyssey/gopickle) - Go 语言 Python pickle 解码器
- [OpenTelemetry](https://opentelemetry.io/) - 可观测性框架
