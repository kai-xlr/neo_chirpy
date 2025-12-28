# NeoChirpy

A lightweight HTTP server written in Go with built-in metrics tracking.

## Features

- **Static File Serving**: Serves HTML, CSS, and assets from the root directory
- **Request Metrics**: Tracks the number of requests to `/app/*` endpoints
- **Health Check**: Provides a readiness endpoint for monitoring
- **Metrics Dashboard**: View request statistics
- **Metrics Reset**: Clear the request counter

## Endpoints

- `GET /` - Root file server
- `GET /app/*` - File server with request tracking
- `GET /healthz` - Health check endpoint (returns "OK")
- `GET /metrics` - Display hit counter
- `POST /reset` - Reset hit counter to 0

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
├── index.html      # Landing page
└── assets/         # Static assets (images, etc.)
```

## Architecture

- **Thread-Safe Metrics**: Uses `atomic.Int32` for concurrent request counting
- **Middleware Pattern**: Request tracking implemented as HTTP middleware
- **Standard Library**: Built entirely with Go's standard `net/http` package
