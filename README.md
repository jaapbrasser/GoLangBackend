# GoLangBackend

A Go-based backend service built with the Gin framework that checks the existence of GitHub repositories.

## Overview

This service provides a simple API endpoint to check if a GitHub repository exists by owner and repository name. It follows a clean three-layer architecture (Handler, Service, Integration) for maintainability and testability.

## Features

- ✅ Check GitHub repository existence
- ✅ RESTful API with JSON responses
- ✅ Structured logging with Zap
- ✅ Environment-based configuration
- ✅ Docker-ready
- ✅ Comprehensive test suite
- ✅ Standardized error handling
- ✅ Health check endpoint

## Tech Stack

- **Language**: Go 1.25+
- **HTTP Framework**: Gin
- **Configuration**: Viper
- **Logging**: Zap
- **HTTP Client**: Standard library `net/http`
- **Testing**: Testify + standard library `testing`
- **Environment**: `.env` file support

## Getting Started

### Prerequisites

- Go 1.25+ installed
- Git (for cloning)
- Internet access (to reach GitHub API)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/GoLangBackend.git
   cd GoLangBackend
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Create a `.env` file (optional, uses defaults if not present):
   ```bash
   cp .env.example .env
   # Edit .env if you want to change PORT or ENVIRONMENT
   ```

### Running the Application

#### Development Mode

```bash
go run ./cmd/api
```

The server will start on port 8080 (or as specified in your `.env` file).

#### Using Docker

Build the Docker image:
```bash
docker build -to golangbackend .
```

Run the container:
```bash
docker run -p 8080:8080 golangbackend
```

## API Endpoints

All API endpoints are prefixed with `/api/v1` unless otherwise noted.

### Health Check
- **GET** `/health`
- Returns 200 OK if the service is running
- **Response**:
  ```json
  {
    "status": "ok"
  }
  ```
- **Example curl command**:
  ```bash
  curl -X GET http://localhost:8080/health
  ```
  Expected response:
  ```json
  {"status":"ok"}
  ```

### Check Repository Existence
- **POST** `/api/v1/repositories/check`
- **Request Body**:
  ```json
  {
    "owner": "string",
    "repo": "string"
  }
  ```
- **Success Response** (200 OK):
  ```json
  {
    "success": true,
    "data": {
      "exists": true,
      "htmlUrl": "https://github.com/owner/repo"
    }
  }
  ```
- **Repository Not Found** (200 OK with exists=false):
  ```json
  {
    "success": true,
    "data": {
      "exists": false,
      "htmlUrl": ""
    }
  }
  ```
- **Error Response** (4xx/5xx):
  ```json
  {
    "success": false,
    "error": "error message"
  }
  ```
- **Example curl commands**:
  ```bash
  # Check for existing repository
  curl -X POST http://localhost:8080/api/v1/repositories/check \
    -H "Content-Type: application/json" \
    -d '{"owner":"golang","repo":"go"}'
  ```
  Expected response:
  ```json
  {"success":true,"data":{"exists":true,"htmlUrl":"https://github.com/golang/go"}}
  ```
  
  ```bash
  # Check for non-existing repository
  curl -X POST http://localhost:8080/api/v1/repositories/check \
    -H "Content-Type: application/json" \
    -d '{"owner":"nonexistentuser","repo":"nonexistentrepo123"}'
  ```
  Expected response:
  ```json
  {"success":true,"data":{"exists":false,"htmlUrl":""}}
  ```

## Common Use Cases

1. **Repository Validation**: Before performing operations on a GitHub repository, validate that it exists.
2. **CI/CD Pipeline**: Integrate into deployment pipelines to verify repository existence before triggering builds.
3. **Dependency Checking**: Verify that dependent repositories exist in microservices architectures.
4. **Repository Discovery**: Build tools that need to check the availability of repositories before cloning or fetching.

## Testing

Run the test suite:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test ./... -cover
```

## Configuration

The service uses Viper for configuration, which reads from:
1. Environment variables
2. `.env` file in the root directory

Available configuration options:
- `PORT`: Port to run the server on (default: 8080)
- `ENVIRONMENT`: Application environment (default: development)

## Project Structure

```
GoLangBackend/
├── cmd/api/main.go          # Application entry point
├── internal/
│   ├── config/              # Configuration loading
│   ├── dto/                 # Data Transfer Objects
│   ├── handler/             # HTTP handlers
│   ├── middleware/          # Gin middleware
│   ├── model/               # Domain models
│   ├── router/              # Route setup
│   └── service/             # Business logic and external integrations
├── pkg/
│   ├── errors/              # Custom error types
│   ├── logger/              # Logger initialization
│   └── response/            # Standardized response helpers
├── .env.example             # Example environment file
├── go.mod                   # Go dependencies
├── go.sum                   # Go dependency lockfile
└── LICENSE                  # MIT License
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please make sure to update tests as appropriate.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [Zap Logger](https://github.com/uber-go/zap)
- [Viper Configuration](https://github.com/spf13/viper)
- [Testify](https://github.com/stretchr/testify)