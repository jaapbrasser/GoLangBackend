# Go API Architecture — Gin Framework

## Overview

This document outlines the architecture of a Go-based service built with the [Gin](https://github.com/gin-gonic/gin) web framework that checks the existence of GitHub repositories. It covers project structure, layered architecture, middleware strategy, error handling, and deployment considerations.

---

## Tech Stack

| Concern             | Tool / Library                          |
|---------------------|-----------------------------------------|
| Language            | Go (1.25+)                              |
| HTTP Framework      | [Gin](https://github.com/gin-gonic/gin) |
| Configuration       | [Viper](https://github.com/spf13/viper) |
| Validation          | Gin built-in (`binding` struct tags)    |
| Logging             | [Zap](https://github.com/uber-go/zap)   |
| Testing             | `testing` stdlib + [testify](https://github.com/stretchr/testify) |

---

## Project Structure

```
GoLangBackend/
├── cmd/
│   └── api/
│       └── main.go              # Entry point — wires everything together
├── internal/
│   ├── config/
│   │   └── config.go            # App configuration (env vars, Viper)
│   ├── dto/
│   │   └── repository_dto.go    # Request/Response DTOs for repository operations
│   ├── handler/
│   │   ├── health_handler.go    # Health check handler
│   │   └── repository_handler.go # Handler for repository operations
│   ├── middleware/
│   │   └── logger.go            # Request logging middleware
│   ├── model/
│   │   └── repository.go        # Domain models (structs)
│   ├── router/
│   │   └── router.go            # Route setup
│   └── service/
│       ├── github_service.go    # Business logic for GitHub operations
│       └── github_service_test.go # Tests for GitHub service
├── pkg/
│   ├── errors/
│   │   └── errors.go            # Shared error types and helpers
│   ├── logger/
│   │   └── logger.go            # Logger initialization and helpers
│   └── response/
│       └── response.go          # Standardised JSON response helpers
├── .env.example
├── go.mod
├── go.sum
└── LICENSE
```

> **Note:** The `internal/` directory is a Go convention — packages inside it cannot be imported by external modules, enforcing encapsulation.

---

## Layered Architecture

The API follows a clean three-layer architecture. Each layer has a single responsibility and only communicates with the layer directly below it.

```
Request
    │
    ▼
┌──────────────┐
│   Handler    │  ← Parses HTTP request, calls service, writes response
└──────┬───────┘
       │
       ▼
┌──────────────┐
│   Service    │  ← Business logic, validation, orchestration
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  Integration │  ← External API calls (no logic, just data access)
└──────────────┘
```

### Handler
- Binds and validates incoming JSON/query params using Gin's `ShouldBindJSON`
- Calls the service layer
- Writes a standardised JSON response
- Contains **no** business logic

### Service
- Implements all business rules
- Calls one or more integrations
- Returns domain models or errors
- Is independently unit-testable (no HTTP context)

### Integration
- Executes external API calls via HTTP clients
- Returns raw domain models
- Contains **no** business logic

---

## Routing

Define all routes in a dedicated `router.go` file, grouped by resource and version.

```go
func SetupRouter() *gin.Engine {
    r := gin.New()

    r.Use(middleware.Logger())
    r.Use(gin.Recovery())

    r.GET("/health", handler.Health)

    repoService := service.NewGitHubService()
    repoHandler := handler.NewRepositoryHandler(repoService)

    v1 := r.Group("/api/v1")
    {
        v1.POST("/repositories/check", repoHandler.CheckRepository)
    }

    return r
}
```

---

## Middleware

Apply middleware at three levels: globally, per group, or per route.

| Middleware     | Scope  | Purpose                                       |
|----------------|--------|-----------------------------------------------|
| Logger         | Global | Log method, path, status, and latency         |
| Recovery       | Global | Recover from panics, return 500               |

---

## Standardised JSON Responses

Use a consistent response envelope across all endpoints.

```go
// pkg/response/response.go

type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{Success: true, Data: data})
}

func Fail(c *gin.Context, status int, message string) {
    c.JSON(status, Response{Success: false, Error: message})
}
```

**Success:**
```json
{
  "success": true,
  "data": { "id": 1, "name": "Alice" }
}
```

**Error:**
```json
{
  "success": false,
  "error": "user not found"
}
```

---

## Error Handling

Define sentinel errors in a shared package and map them to HTTP status codes in the handler layer.

```go
// pkg/errors/errors.go
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrConflict     = errors.New("conflict")
)
```

```go
// handler — map service errors to HTTP codes
user, err := h.service.GetUser(id)
if errors.Is(err, apperrors.ErrNotFound) {
    response.Fail(c, http.StatusNotFound, "user not found")
    return
}
if err != nil {
    response.Fail(c, http.StatusInternalServerError, "internal error")
    return
}
response.OK(c, user)
```

---

## Configuration

Use environment variables for all configuration. Load them with Viper so they can be sourced from a `.env` file in development and real env vars in production.

```go
// internal/config/config.go
type Config struct {
    Port        string
    Environment string // "development" | "production"
}
```

**.env.example**
```
PORT=8080
ENVIRONMENT=development
```

---

## Entry Point

```go
// cmd/api/main.go
func main() {
    cfg := config.Load()

    db := database.Connect(cfg.DatabaseURL)
    
    userRepo    := repository.NewUserRepository(db)
    userService := service.NewUserService(userRepo)
    userHandler := handler.NewUserHandler(userService)

    router := SetupRouter(userHandler, middleware.Auth(cfg.JWTSecret))

    log.Printf("Starting server on :%s", cfg.Port)
    if err := router.Run(":" + cfg.Port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

---

## Testing Strategy

| Layer      | Test Type    | Tool                         |
|------------|--------------|------------------------------|
| Handler    | Integration  | `net/http/httptest` + testify |
| Service    | Unit         | testify/mock                  |
| Integration| Unit         | testify/mock (for HTTP client)|

Use interfaces for the service and integration layers to enable easy mocking:

```go
type GitHubService interface {
    CheckRepositoryExists(owner, repo string) (*model.Repository, error)
}
```

---

## Dockerfile

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/api

# Run stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

---

## Key Principles

- **Thin handlers** — handlers should only parse input and write output; all logic lives in services.
- **Dependency injection** — pass dependencies (services) via constructors, not global vars.
- **Interfaces at boundaries** — define interfaces for services and integrations to keep layers decoupled and testable.
- **Explicit error handling** — always handle errors; never silently ignore `err`.
- **Environment-based config** — no hardcoded secrets or URLs; use `.env` locally and real env vars in production.
- **Version your API** — prefix all routes with `/api/v1` from day one to enable future non-breaking changes.
