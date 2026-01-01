# NeoChirpy

A lightweight HTTP server written in Go with built-in metrics tracking, chirp validation, and PostgreSQL database integration.

## Features

- **Static File Serving**: Serves HTML, CSS, and assets from the root directory
- **Request Metrics**: Tracks the number of requests to `/app/*` endpoints
- **Health Check**: Provides a readiness endpoint for monitoring
- **Chirp Validation**: Validates chirp messages (max 140 characters)
- **Profanity Filtering**: Automatically sanitizes banned words in chirps
- **Database Integration**: PostgreSQL database with user management
- **Metrics Dashboard**: View request statistics in HTML format
- **Metrics Reset**: Clear the request counter

## Endpoints

### Public
- `GET /` - Root file server
- `GET /app/*` - File server with request tracking

### API
- `GET /api/healthz` - Health check endpoint (returns "OK")
- `POST /api/validate_chirp` - Validate and sanitize chirp message (max 140 characters, filters profanity)

### Admin
- `GET /admin/metrics` - Display hit counter with HTML dashboard
- `POST /admin/reset` - Reset hit counter to 0

All endpoints return 405 (Method Not Allowed) for unsupported HTTP methods.

## Getting Started

### Prerequisites

- Go 1.25.2 or higher
- PostgreSQL database
- Environment variables configured (see Configuration)

### Running the Server

```bash
go build -o out
./out
```

The server will start on port 8080.

### Configuration

Create a `.env` file in the project root:

```env
DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
```

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
├── sanitize.go     # Profanity filtering logic
├── sql/
│   └── schema/     # Database migration files
├── internal/
│   └── database/   # Generated database access code
├── index.html      # Landing page
└── assets/         # Static assets (images, etc.)
```

## Architecture

- **Thread-Safe Metrics**: Uses `atomic.Int32` for concurrent request counting
- **Middleware Pattern**: Request tracking implemented as HTTP middleware
- **JSON API**: Structured error handling and JSON responses
- **Code Organization**: Modular main function with `initDatabase()`, `setupRouter()`, and `startServer()` helpers
- **Database Layer**: PostgreSQL with sqlc-generated type-safe queries
- **Migration Management**: Goose for database schema versioning
