# NeoChirpy

A lightweight HTTP server written in Go with built-in metrics tracking, chirp validation, and PostgreSQL database integration.

## Features

- **Static File Serving**: Serves HTML, CSS, and assets from the root directory
- **Request Metrics**: Tracks the number of requests to `/app/*` endpoints
- **Health Check**: Provides a readiness endpoint for monitoring
- **Chirp Management**: Create and store chirp messages (max 140 characters)
- **Profanity Filtering**: Automatically sanitizes banned words in chirps
- **Database Integration**: PostgreSQL database with user and chirp management
- **Metrics Dashboard**: View request statistics in HTML format
- **Metrics Reset**: Clear the request counter

## Endpoints

### Public
- `GET /` - Root file server
- `GET /app/*` - File server with request tracking

### API
- `GET /api/healthz` - Health check endpoint (returns "OK")
- `GET /api/chirps` - Retrieve all chirps (ordered by creation date, oldest first)
- `POST /api/chirps` - Create a new chirp (max 140 characters, filters profanity)
- `POST /api/users` - Create a new user account

### Admin
- `GET /admin/metrics` - Display hit counter with HTML dashboard
- `POST /admin/reset` - Reset hit counter and database (dev environment only)

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
PLATFORM=dev
```

### Development

```bash
go run .
```

### Testing

```bash
# Run all tests
go test

# Run tests with verbose output
go test -v

# Run specific test function
go test -run TestValidateChirpBody

# Run tests with coverage
go test -cover
```

## Project Structure

```
.
├── main.go            # Server setup and configuration
├── types.go           # Request/response type definitions
├── constants.go       # Application-wide constants
├── validation.go      # Input validation functions
├── validation_test.go # Unit tests for validation
├── handlers.go        # HTTP helper functions
├── handlers_api.go    # API endpoint handlers (with documentation)
├── handlers_admin.go  # Admin endpoint handlers
├── handlers_users.go  # User management handlers
├── middleware.go      # HTTP middleware functions
├── json.go            # JSON response helpers and response builders
├── sanitize.go        # Profanity filtering logic
├── AGENTS.md          # Guidelines for agentic coding agents
├── sql/
│   ├── schema/        # Database migration files
│   └── queries/       # SQL queries for sqlc
├── internal/
│   └── database/      # Generated database access code
├── index.html         # Landing page
└── assets/            # Static assets (images, etc.)
```

## Architecture

- **Thread-Safe Metrics**: Uses `atomic.Int32` for concurrent request counting
- **Middleware Pattern**: Request tracking implemented as HTTP middleware
- **JSON API**: Structured error handling and JSON responses
- **Code Organization**: 
  - Main function split into `initDatabase()`, `setupRouter()`, and `startServer()` helpers
  - Handlers organized by domain (API, Admin, Users)
  - Consistent file structure: package → imports → structs → functions
  - Centralized validation with reusable functions
  - Comprehensive documentation and response builders
  - Error handling with standardized messages
- **Input Validation**: Dedicated validation package with error constants
- **Testing**: Unit tests for validation logic
- **Database Layer**: PostgreSQL with sqlc-generated type-safe queries
- **Migration Management**: Goose for database schema versioning
