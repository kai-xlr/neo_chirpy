# AGENTS.md

This file contains guidelines and commands for agentic coding agents working in this Go codebase.

## Project Overview

This is a Go web application called "Chirpy" - a Twitter-like microblogging platform built with:
- Standard library `net/http` router
- PostgreSQL database with sqlc for type-safe queries
- Minimal dependencies (UUID, environment loading, PostgreSQL driver)
- Clean architecture with domain-driven file organization

## Build & Development Commands

### Build & Run
```bash
# Build the application
go build -o out

# Run the built application
./out

# Development mode (build and run)
go run .
```

### Testing Commands
```bash
# Run all tests
go test

# Run tests with verbose output
go test -v

# Run specific test function
go test -run TestValidateChirpBody

# Run tests in specific file
go test ./validation_test.go

# Run tests with coverage
go test -cover

# Generate coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Database Operations
```bash
# Generate database code from SQL (requires sqlc)
sqlc generate

# Database migrations are in sql/schema/ directory
# Migration files follow Goose naming convention: 001_users.sql, 002_chirps.sql
```

## Code Style Guidelines

### Import Organization
```go
import (
    // Standard library first
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"
    "sync/atomic"

    // Third-party packages second
    "github.com/joho/godotenv"
    "github.com/kai-xlr/neo_chirpy/internal/database"
    _ "github.com/lib/pq"
)
```

### Naming Conventions
- **Package names**: lowercase, single word (`main`, `database`)
- **Functions**: camelCase with descriptive names (`ValidateChirpBody`, `handlerReadiness`)
- **Variables**: camelCase (`fileserverHits`, `dbQueries`)
- **Constants**: PascalCase with descriptive prefixes (`MaxChirpLength`, `ContentTypeJSON`)
- **Types**: PascalCase (`apiConfig`, `chirpRequest`, `errorResponse`)
- **Error variables**: `Err` prefix (`ErrChirpTooLong`, `ErrEmailInvalid`)

### Type Definitions
```go
type chirpCreateRequest struct {
    Body   string    `json:"body"`
    UserID uuid.UUID `json:"user_id"`
}

type chirpCreateResponse struct {
    ID        uuid.UUID `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    UserID    uuid.UUID `json:"user_id"`
    Body      string    `json:"body"`
}
```

### Error Handling
- Use centralized error variables defined in `validation.go`
- Follow consistent error response format via `respondWithError()`
- Log errors appropriately, especially 5XX responses
- Use early return pattern for error conditions

```go
// Centralized error variables
var (
    ErrChirpTooLong   = errors.New("Chirp is too long")
    ErrChirpEmpty     = errors.New("Chirp cannot be empty")
    ErrEmailInvalid   = errors.New("Invalid email address")
    ErrEmailEmpty     = errors.New("Email cannot be empty")
    ErrUserIDInvalid  = errors.New("Invalid user ID")
)

// Error response pattern
if err != nil {
    respondWithError(w, http.StatusBadRequest, err.Error())
    return
}
```

## File Organization

### Core Application Files
- `main.go` - Application entry point and server setup
- `types.go` - Request/response type definitions
- `constants.go` - Application-wide constants
- `validation.go` - Input validation functions and error variables
- `validation_test.go` - Unit tests for validation

### HTTP Layer
- `handlers.go` - HTTP helper functions and middleware
- `handlers_api.go` - General API endpoint handlers
- `handlers_admin.go` - Admin-specific endpoint handlers
- `handlers_users.go` - User management handlers
- `middleware.go` - HTTP middleware functions
- `json.go` - JSON response helpers
- `sanitize.go` - Profanity filtering logic

### Database Layer
- `sqlc.yaml` - SQLC configuration for code generation
- `sql/schema/` - Database migration files (Goose format)
- `sql/queries/` - SQL queries for sqlc generation
- `internal/database/` - Generated database access code

### Configuration
- `.env` - Environment variables (DB_URL, PLATFORM)
- `go.mod`, `go.sum` - Go modules and dependencies
- `.gitignore` - Git ignore rules

## Testing Guidelines

### Test Structure
Use table-driven tests with comprehensive test cases:

```go
func TestValidateChirpBody(t *testing.T) {
    tests := []struct {
        name    string
        body    string
        wantErr error
    }{
        {
            name:    "valid chirp",
            body:    "This is a valid chirp",
            wantErr: nil,
        },
        {
            name:    "empty chirp",
            body:    "",
            wantErr: ErrChirpEmpty,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateChirpBody(tt.body)
            if err != tt.wantErr {
                t.Errorf("ValidateChirpBody() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Testing Best Practices
- Test functions named `Test[FunctionName]`
- Use subtests with `t.Run()` for different scenarios
- Compare errors using predefined error variables
- Test edge cases (empty input, max length, invalid formats)
- Aim for comprehensive coverage of validation logic

## Architecture Patterns

### Database Layer
- PostgreSQL with sqlc-generated type-safe queries
- Migration files using Goose format in `sql/schema/`
- Models in `internal/database` package
- Context-aware database operations

### HTTP Layer
- Standard library `net/http` router
- Middleware pattern for cross-cutting concerns
- JSON API with consistent error handling
- Method validation using helper functions like `requireMethod()`

### Concurrency
- Use `atomic.Int32` for thread-safe metrics
- Propagate context for database operations
- Rely on HTTP server's built-in concurrency

## Dependencies

### Core Dependencies
- `github.com/google/uuid v1.6.0` - UUID generation
- `github.com/joho/godotenv v1.5.1` - Environment variable loading
- `github.com/lib/pq v1.10.9` - PostgreSQL driver

### Development Tools
- `sqlc` - Type-safe SQL code generation
- Standard Go testing framework

## Key Conventions

1. **File Organization**: One concern per file with clear naming
2. **Error Handling**: Centralized error variables, consistent responses
3. **Testing**: Table-driven tests with comprehensive coverage
4. **Database**: Type-safe queries with sqlc, migration-first approach
5. **HTTP**: Standard library with middleware pattern
6. **Configuration**: Environment-based, .env for local development
7. **Dependencies**: Minimal, well-chosen third-party libraries
8. **Documentation**: Comprehensive README with clear structure

## Important Notes

- This codebase follows Go best practices with clean architecture
- Always run `go test` before committing changes
- Use sqlc for database operations - never write raw SQL in application code
- Follow the established error handling patterns with centralized error variables
- Maintain the import organization and naming conventions
- Test validation logic thoroughly using table-driven tests