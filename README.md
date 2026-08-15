#  — Distributed Rate Limiter

A beginner-friendly Go project for building a distributed rate limiter from scratch.

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go          # Application entry point — starts the HTTP server
├── internal/
│   ├── handler/             # HTTP handler functions for API endpoints
│   ├── middleware/           # HTTP middleware (rate limiting, logging, etc.)
│   ├── limiter/             # Core rate limiting algorithms and data structures
│   ├── service/             # Business logic layer between handlers and limiter
│   └── config/              # Application configuration (ports, limits, env)
├── go.mod                   # Go module definition
└── README.md
```

### What each folder does

| Folder | Responsibility |
|---|---|
| `cmd/server/` | Entry point. Wires everything together and starts the HTTP server. |
| `internal/handler/` | HTTP handlers — parse requests, call services, write responses. |
| `internal/middleware/` | Middleware that wraps handlers (rate limiting will live here). |
| `internal/limiter/` | The actual rate limiting logic (token bucket, sliding window, etc.). |
| `internal/service/` | Business logic that coordinates between handlers and the limiter. |
| `internal/config/` | Loads and holds configuration values (server port, rate limits, etc.). |

> The `internal/` directory is a Go convention — packages inside it **cannot** be imported by external projects. This enforces encapsulation.

## Running

```bash
go run ./cmd/server
```

The server starts on **port 8080**. Test the health endpoint:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

## What's Next

- [ ] Add configuration loading (`internal/config`)
- [ ] Implement a token bucket rate limiter (`internal/limiter`)
- [ ] Create rate limiting middleware (`internal/middleware`)
- [ ] Add more API endpoints (`internal/handler`)
- [ ] Add Redis for distributed state
- [ ] Add tests
