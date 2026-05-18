# Go API Architecture — Gin Framework

## Overview

This document outlines the recommended architecture for building a production-ready REST API using Go and the [Gin](https://github.com/gin-gonic/gin) web framework. It covers project structure, layered architecture, middleware strategy, error handling, and deployment considerations.

---

## Tech Stack

| Concern             | Tool / Library                          |
|---------------------|-----------------------------------------|
| Language            | Go (1.21+)                              |
| HTTP Framework      | [Gin](https://github.com/gin-gonic/gin) |
| ORM / DB Layer      | [GORM](https://gorm.io/) or `database/sql` |
| Configuration       | [Viper](https://github.com/spf13/viper) |
| Validation          | Gin built-in (`binding` struct tags)    |
| Logging             | [Zap](https://github.com/uber-go/zap) or `slog` (stdlib) |
| Authentication      | JWT via [golang-jwt](https://github.com/golang-jwt/jwt) |
| Testing             | `testing` stdlib + [testify](https://github.com/stretchr/testify) |
| Containerisation    | Docker + Docker Compose                 |

---

## Project Structure

```
my-api/
├── cmd/
│   └── api/
│       └── main.go              # Entry point — wires everything together
├── internal/
│   ├── config/
│   │   └── config.go            # App configuration (env vars, Viper)
│   ├── handler/
│   │   ├── user_handler.go      # HTTP handlers (thin layer — no business logic)
│   │   └── health_handler.go
│   ├── middleware/
│   │   ├── auth.go              # JWT authentication middleware
│   │   ├── logger.go            # Request logging middleware
│   │   └── cors.go              # CORS middleware
│   ├── service/
│   │   └── user_service.go      # Business logic layer
│   ├── repository/
│   │   └── user_repository.go   # Database access layer
│   ├── model/
│   │   └── user.go              # Domain models (structs)
│   └── dto/
│       └── user_dto.go          # Request/Response DTOs (data transfer objects)
├── pkg/
│   ├── errors/
│   │   └── errors.go            # Shared error types and helpers
│   └── response/
│       └── response.go          # Standardised JSON response helpers
├── migrations/
│   └── 001_create_users.sql     # Database migration files
├── .env.example
├── docker-compose.yml
├── Dockerfile
└── go.mod
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
│  Repository  │  ← Database queries (no logic, just data access)
└──────────────┘
```

### Handler
- Binds and validates incoming JSON/query params using Gin's `ShouldBindJSON`
- Calls the service layer
- Writes a standardised JSON response
- Contains **no** business logic

### Service
- Implements all business rules
- Calls one or more repositories
- Returns domain models or errors
- Is independently unit-testable (no HTTP context)

### Repository
- Executes database queries via GORM or `database/sql`
- Returns raw domain models
- Contains **no** business logic

---

## Routing

Define all routes in `main.go` or a dedicated `router.go` file, grouped by resource and version.

```go
func SetupRouter(
    userHandler *handler.UserHandler,
    authMiddleware gin.HandlerFunc,
) *gin.Engine {
    r := gin.New()

    // Global middleware
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())
    r.Use(middleware.CORS())

    // Health check (no auth)
    r.GET("/health", handler.Health)

    // API v1
    v1 := r.Group("/api/v1")
    {
        // Public routes
        auth := v1.Group("/auth")
        {
            auth.POST("/register", userHandler.Register)
            auth.POST("/login", userHandler.Login)
        }

        // Protected routes
        users := v1.Group("/users")
        users.Use(authMiddleware)
        {
            users.GET("", userHandler.ListUsers)
            users.GET("/:id", userHandler.GetUser)
            users.PUT("/:id", userHandler.UpdateUser)
            users.DELETE("/:id", userHandler.DeleteUser)
        }
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
| CORS           | Global | Set `Access-Control-Allow-*` headers          |
| Auth (JWT)     | Group  | Validate Bearer token, inject user into ctx   |
| RateLimiter    | Group  | Throttle requests per IP or API key           |

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
    DatabaseURL string
    JWTSecret   string
    Environment string // "development" | "production"
}
```

**.env.example**
```
PORT=8080
DATABASE_URL=postgres://user:pass@localhost:5432/mydb?sslmode=disable
JWT_SECRET=change-me-in-production
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
| Repository | Integration  | Test DB (Docker) + testify    |

Use interfaces for the service and repository layers to enable easy mocking:

```go
type UserService interface {
    GetUser(id uint) (*model.User, error)
    CreateUser(dto dto.CreateUserRequest) (*model.User, error)
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
- **Dependency injection** — pass dependencies (DB, services) via constructors, not global vars.
- **Interfaces at boundaries** — define interfaces for services and repositories to keep layers decoupled and testable.
- **Explicit error handling** — always handle errors; never silently ignore `err`.
- **Environment-based config** — no hardcoded secrets or URLs; use `.env` locally and real env vars in production.
- **Version your API** — prefix all routes with `/api/v1` from day one to enable future non-breaking changes.
