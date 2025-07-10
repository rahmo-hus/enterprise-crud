# Enterprise CRUD API

A RESTful API for user management built with Go, Gin, GORM, and PostgreSQL.

## Features

- User registration and retrieval
- REST endpoints with proper HTTP status codes
- Clean architecture with domain-driven design
- Comprehensive test coverage
- Database migrations
- Password hashing with bcrypt
- JSON validation
- **Swagger/OpenAPI documentation**
- Interactive API testing interface
- Extensive inline documentation

## Project Structure

```
enterprise-crud/
├── cmd/
│   └── migrate/           # Database migration tool
├── docs/                  # Swagger documentation files
├── internal/
│   ├── domain/
│   │   └── user/          # User domain logic and interfaces
│   ├── dto/
│   │   └── user/          # Data transfer objects
│   ├── infrastructure/
│   │   └── database/      # Database implementations
│   └── presentation/
│       └── http/          # HTTP handlers
├── migrations/            # SQL migration files
├── docker-compose.yml     # PostgreSQL database setup
├── .env                   # Environment variables
└── main.go               # Application entry point
```

## Quick Start

### 1. Start Database

```bash
# Start PostgreSQL database
docker-compose up -d

# Verify database is running
docker-compose ps
```

### 2. Run Migrations

```bash
# Run database migrations
go run cmd/migrate/main.go up

# Check migration status
go run cmd/migrate/main.go version
```

### 3. Start API Server

```bash
# Start the server
go run main.go

# Server will start on port 8080
# API available at: http://localhost:8080/api/v1
# Swagger UI available at: http://localhost:8080/swagger/index.html
```

## API Documentation

### Swagger UI
🚀 **Interactive API Documentation**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

The Swagger UI provides:
- Complete API specification
- Interactive endpoint testing
- Request/response examples
- Schema definitions
- Authentication details

### Raw Documentation
- **JSON**: [http://localhost:8080/swagger/doc.json](http://localhost:8080/swagger/doc.json)
- **YAML**: Available in `docs/swagger.yaml`

## API Endpoints

### Health Check
```
GET /health
```

### Authentication

#### Login
```
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Success Response (200):**
```json
{
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "username": "testuser"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": 1735689600
}
```

**Error Responses:**
- `400` - Invalid request body or validation errors
- `401` - Invalid email or password
- `500` - Internal server error

### User Management

#### Create User (Public)
```
POST /api/v1/users
Content-Type: application/json

{
  "email": "user@example.com",
  "username": "testuser",
  "password": "password123"
}
```

**Success Response (201):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "username": "testuser"
}
```

**Error Responses:**
- `400` - Invalid request body or validation errors
- `409` - User with email already exists
- `500` - Internal server error

#### Get User by Email (Public)
```
GET /api/v1/users/{email}
```

**Success Response (200):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "username": "testuser"
}
```

**Error Responses:**
- `400` - Invalid email parameter
- `404` - User not found
- `500` - Internal server error

#### Get User Profile (Protected)
```
GET /api/v1/users/profile
Authorization: Bearer <JWT_TOKEN>
```

**Success Response (200):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "username": "testuser"
}
```

**Error Responses:**
- `401` - Unauthorized - invalid or missing token
- `500` - Internal server error

## Testing

### Run All Tests
```bash
# Run all tests with coverage
go test ./... -v -cover

# Run tests with detailed coverage report
go test ./... -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run Specific Test Suites
```bash
# Test user service
go test ./internal/domain/user/ -v

# Test user handlers
go test ./internal/presentation/http/ -v

# Test user repository
go test ./internal/infrastructure/database/ -v
```

## Database Operations

### Migration Commands
```bash
# Apply migrations
go run cmd/migrate/main.go up

# Rollback migrations
go run cmd/migrate/main.go down

# Check migration version
go run cmd/migrate/main.go version

# Force specific version
go run cmd/migrate/main.go force 1
```

### Database Schema

**Users Table:**
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## Configuration

### Environment Variables

```bash
# Database Configuration
DATABASE_URL=postgres://postgres:postgres@localhost:5433/enterprise_crud?sslmode=disable
DB_HOST=localhost
DB_PORT=5433
DB_NAME=enterprise_crud
DB_USER=postgres
DB_PASSWORD=postgres

# Server Configuration
PORT=8080

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_ISSUER=enterprise-crud-api
JWT_EXPIRATION_HOURS=720
```

### Example API Calls with JWT

#### 1. Create a User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "john_doe",
    "password": "password123"
  }'
```

#### 2. Login to Get Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

**Save the token from the response:**
```json
{
  "user": {...},
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": 1735689600
}
```

#### 3. Access Protected Endpoint
```bash
# Replace <TOKEN> with the actual JWT token from login
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer <TOKEN>"
```

#### 4. Test Token Expiration
The JWT tokens are long-lived (30 days by default). They will automatically expire and require re-authentication.

### Docker Compose Configuration

The provided `docker-compose.yml` sets up:
- PostgreSQL 15 Alpine
- Database: `enterprise_crud`
- Username/Password: `postgres/postgres`
- Port: `5433`
- Data persistence with named volume

## Development

### Code Style

The codebase follows Go best practices with:
- Clean architecture principles
- Extensive inline documentation
- Proper error handling
- Input validation
- Test-driven development

### Architecture Layers

1. **Domain Layer** (`internal/domain/`):
   - Business logic and interfaces
   - Domain entities and services
   - Repository interfaces

2. **Infrastructure Layer** (`internal/infrastructure/`):
   - Database implementations
   - External service integrations
   - Repository implementations

3. **Presentation Layer** (`internal/presentation/`):
   - HTTP handlers
   - Request/response mapping
   - Route registration

4. **DTO Layer** (`internal/dto/`):
   - Data transfer objects
   - Request/response structures
   - Validation rules

### Adding New Features

1. Define domain interfaces in `internal/domain/`
2. Create DTOs in `internal/dto/`
3. Implement infrastructure in `internal/infrastructure/`
4. Add HTTP handlers in `internal/presentation/`
5. Write comprehensive tests for all layers
6. Update documentation

## Troubleshooting

### Database Connection Issues
```bash
# Check if database is running
docker-compose ps

# Check database logs
docker-compose logs db

# Restart database
docker-compose restart db
```

### Migration Issues
```bash
# Check current migration status
go run cmd/migrate/main.go version

# Force migration to specific version
go run cmd/migrate/main.go force 1

# Rollback and reapply
go run cmd/migrate/main.go down
go run cmd/migrate/main.go up
```

### Port Conflicts
```bash
# Change port in docker-compose.yml or .env
PORT=8081 go run main.go
```

## Contributing

1. Follow the existing code style and documentation patterns
2. Add comprehensive tests for new features
3. Update documentation when adding new endpoints
4. Ensure all tests pass before submitting changes

## License

This project is for educational/demonstration purposes.