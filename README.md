# NeoChirpy

A lightweight HTTP server written in Go with built-in metrics tracking, chirp validation, user authentication, and PostgreSQL database integration.

## Features

- **Static File Serving**: Serves HTML, CSS, and assets from the root directory
- **Request Metrics**: Tracks the number of requests to `/app/*` endpoints
- **Health Check**: Provides a readiness endpoint for monitoring
- **Chirp Management**: Create, retrieve, and store chirp messages (max 140 characters)
- **Profanity Filtering**: Automatically sanitizes banned words in chirps
- **Individual Chirp Retrieval**: Fetch specific chirps by UUID
- **Advanced Chirp Filtering**: Filter chirps by author ID and sort by creation date (asc/desc)
- **User Authentication**: Secure password-based user registration and login
- **Password Security**: Argon2id hashing for secure password storage
- **JWT Authentication**: Complete JWT token generation and validation with HS256 signing
- **Bearer Token Authentication**: Protected endpoints require valid JWT in Authorization header
- **Configurable Token Expiration**: Optional custom expiration times for JWT tokens
- **Database Integration**: PostgreSQL database with user and chirp management
- **Metrics Dashboard**: View request statistics in HTML format
- **Metrics Reset**: Clear the request counter

## Endpoints

### Public
- `GET /` - Root file server
- `GET /app/*` - File server with request tracking

### API
- `GET /api/healthz` - Health check endpoint (returns "OK")
- `GET /api/chirps` - Retrieve chirps with optional filtering and sorting
- `GET /api/chirps/{id}` - Retrieve a specific chirp by ID
- `POST /api/chirps` - Create a new chirp (requires authentication, max 140 characters, filters profanity)
- `POST /api/users` - Create a new user account with password
- `POST /api/login` - Authenticate user and return access token

#### Authentication

**User Registration**
```json
POST /api/users
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**User Login**
```json
POST /api/login
{
  "email": "user@example.com", 
  "password": "securepassword123",
  "expires_in_seconds": 3600
}
```

Returns user data with signed JWT access token for authenticated sessions. The `expires_in_seconds` field is optional (defaults to 1 hour, maximum 1 hour).

**Creating Chirps (Authenticated)**
```json
POST /api/chirps
Authorization: Bearer <jwt_token>
{
  "body": "This is a new chirp!"
}
```

Requires a valid JWT token in the Authorization header. The user ID is automatically extracted from the token.

**Retrieving Chirps**
```bash
GET /api/chirps
```

Supports optional query parameters for filtering and sorting:

- `author_id` (UUID): Filter chirps by specific author
- `sort` (string): Sort order - `asc` (default) or `desc`

Examples:
```bash
# Get all chirps, sorted by creation date (oldest first)
GET /api/chirps

# Get all chirps, sorted by creation date (newest first)  
GET /api/chirps?sort=desc

# Get chirps from specific author, sorted oldest first
GET /api/chirps?author_id=550e8400-e29b-41d4-a716-446655440000

# Get chirps from specific author, sorted newest first
GET /api/chirps?author_id=550e8400-e29b-41d4-a716-446655440000&sort=desc
```

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
# Build and run
go build -o chirpy ./cmd/web
./chirpy

# Or run directly
go run ./cmd/web
```

The server will start on port 8080.

### Configuration

Create a `.env` file in the project root:

```env
DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
JWT_SECRET=<your-super-secret-jwt-key>
```

Generate a secure JWT secret with:
```bash
openssl rand -base64 64
```

### Development

```bash
# Run from project root
go run ./cmd/web

# Or build and run
go build -o chirpy ./cmd/web
./chirpy
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

# Run specific package tests
go test ./internal/auth/ -v
go test ./pkg/validation/ -v

# Run all tests with coverage
go test -cover ./...
```
```

## Project Structure

```
.
├── cmd/
│   └── web/
│       └── main.go            # Application entry point and server setup
├── pkg/                     # Public library code organized by domain
│   ├── admin/
│   │   └── handlers_admin.go # Admin endpoints and metrics
│   ├── chirp/
│   │   ├── handlers.go       # Chirp CRUD operations
│   │   └── sanitize.go     # Profanity filtering
│   ├── handlers/
│   │   ├── handlers.go      # Common HTTP utilities
│   │   └── health.go       # Health check endpoint
│   ├── middleware/
│   │   └── middleware.go   # HTTP middleware components
│   ├── types/
│   │   ├── types.go         # Shared types and structs
│   │   └── constants.go     # Application constants
│   ├── user/
│   │   ├── handlers.go       # User management endpoints
│   │   └── auth_helpers.go  # Authentication helpers
│   ├── validation/
│   │   ├── validation.go     # Input validation logic
│   │   ├── constants.go     # Validation constants
│   │   └── validation_test.go # Unit tests
│   └── webhook/
│       └── handlers.go      # External webhook handling
├── internal/                # Internal packages (not for external use)
│   ├── auth/              # Authentication utilities
│   │   ├── passwords.go    # Password hashing and verification
│   │   └── passwords_test.go # Auth tests
│   └── database/          # Database access layer
│       ├── db.go          # Database connection
│       └── *.sql.go      # Generated queries (sqlc)
├── sql/                   # Database schema and queries
│   ├── schema/           # Migration files (Goose format)
│   └── queries/          # SQL queries for code generation
├── index.html           # Landing page
├── assets/              # Static assets
├── AGENTS.md           # Development guidelines
├── sqlc.yaml          # SQL code generation config
└── go.mod/go.sum      # Go module definition
```

## Architecture

- **Thread-Safe Metrics**: Uses `atomic.Int32` for concurrent request counting
- **Middleware Pattern**: Request tracking implemented as HTTP middleware
- **JSON API**: Structured error handling and JSON responses
- **Authentication System**:
  - Argon2id password hashing for secure storage
  - Complete JWT implementation with HS256 signing
  - Token validation with expiration and signature verification
  - Bearer token extraction from Authorization headers
  - Protected endpoints with automatic user identification
  - Configurable token expiration with security limits
  - Input validation for authentication requests
  - Clear separation between validation and business logic
  - Comprehensive test coverage for JWT operations
- **Code Organization**: 
  - **Idiomatic Go Structure**: Follows standard project-layout conventions
  - **Domain-Driven Design**: Packages organized by business domain (admin, chirp, user, webhook)
  - **Clear Separation**: `cmd/` for applications, `pkg/` for libraries, `internal/` for private code
  - **Clean Architecture**: Each package handles specific responsibilities with minimal coupling
  - **Consistent Structure**: Package → imports → types → functions
  - **Centralized Validation**: Reusable validation functions and error constants
  - **Modular Design**: Each handler package is self-contained and testable
  - **Comprehensive Documentation**: Clear function documentation and README
- **Input Validation**: Dedicated validation package with error constants
- **Testing**: Unit tests for validation logic
- **Database Layer**: PostgreSQL with sqlc-generated type-safe queries
- **Migration Management**: Goose for database schema versioning
- **Security**: Password hashing, input validation, and structured error handling

## Security Features

- **Password Security**: Uses Argon2id (recommended password hashing algorithm)
- **Input Validation**: Comprehensive validation for all user inputs
- **Error Handling**: Consistent error responses that don't leak sensitive information
- **Token Generation**: Complete JWT implementation with proper signing and validation
- **Bearer Token Authentication**: Secure Bearer token extraction and validation
- **Protected Endpoints**: JWT-based authorization for sensitive operations
- **Database Security**: Type-safe SQL queries prevent injection attacks
