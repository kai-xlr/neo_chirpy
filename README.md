# NeoChirpy

A lightweight HTTP server written in Go with built-in metrics tracking and chirp validation.

## Features

- **Static File Serving**: Serves HTML, CSS, and assets from the root directory
- **Request Metrics**: Tracks the number of requests to `/app/*` endpoints
- **Health Check**: Provides a readiness endpoint for monitoring
- **Chirp Validation**: Validates chirp messages (max 140 characters)
- **Metrics Dashboard**: View request statistics in HTML format
- **Metrics Reset**: Clear the request counter

## Endpoints

### Public
- `GET /` - Root file server
- `GET /app/*` - File server with request tracking

### API
- `GET /api/healthz` - Health check endpoint (returns "OK")
- `POST /api/validate_chirp` - Validate chirp message (max 140 characters)

### Admin
- `GET /admin/metrics` - Display hit counter with HTML dashboard
- `POST /admin/reset` - Reset hit counter to 0

All endpoints return 405 (Method Not Allowed) for unsupported HTTP methods.

## Getting Started

### Prerequisites

- Go 1.25.2 or higher

### Running the Server

```bash
go build -o out
./out
```

The server will start on port 8080.

### Development

```bash
go run .
```

## Project Structure

```
.
├── main.go         # Server setup and configuration
├── handlers.go     # HTTP endpoint handlers
├── middleware.go   # HTTP middleware functions
├── json.go         # JSON response helpers
├── index.html      # Landing page
└── assets/         # Static assets (images, etc.)
```

## Architecture

- **Thread-Safe Metrics**: Uses `atomic.Int32` for concurrent request counting
- **Middleware Pattern**: Request tracking implemented as HTTP middleware
- **JSON API**: Structured error handling and JSON responses
- **Standard Library**: Built entirely with Go's standard `net/http` and `encoding/json` packages
