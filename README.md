# Enterprise CRUD API

A comprehensive RESTful API for **event ticketing system** built with Go, Gin, GORM, and PostgreSQL.

## Features

- **User Management**: Registration, authentication, and role-based access control
- **Event Management**: Create, update, and manage events with venues
- **Venue Management**: Full CRUD operations for venue administration
- **Order Management**: Ticket ordering with transaction support
- **JWT Authentication**: Secure API access with role-based permissions
- **Database Transactions**: Atomic operations for ticket purchases
- **Clean Architecture**: Domain-driven design with comprehensive layers
- **Comprehensive Test Coverage**: 80+ unit and integration tests
- **Database Migrations**: Version-controlled schema management
- **Password Security**: bcrypt hashing with salt
- **Input Validation**: JSON schema validation and business rules
- **Swagger/OpenAPI Documentation**: Interactive API testing interface
- **Role-Based Access Control**: USER, ORGANIZER, ADMIN roles

## Project Structure

```
enterprise-crud/
├── cmd/
│   └── migrate/           # Database migration tool
├── docs/                  # Swagger documentation files
├── internal/
│   ├── domain/
│   │   ├── user/          # User domain logic and interfaces
│   │   ├── event/         # Event domain logic and interfaces
│   │   ├── venue/         # Venue domain logic and interfaces
│   │   ├── order/         # Order domain logic and interfaces
│   │   ├── ticket/        # Ticket domain logic and interfaces
│   │   └── role/          # Role domain logic and interfaces
│   ├── dto/
│   │   ├── user/          # User data transfer objects
│   │   ├── event/         # Event data transfer objects
│   │   ├── venue/         # Venue data transfer objects
│   │   └── order/         # Order data transfer objects
│   ├── infrastructure/
│   │   ├── database/      # Database implementations
│   │   └── auth/          # JWT authentication
│   └── presentation/
│       └── http/          # HTTP handlers
├── migrations/            # SQL migration files
├── tests/                 # Integration tests
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

### 🔗 Complete API Documentation
**Interactive Swagger UI**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

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

#### Get User Profile (Protected)
```
GET /api/v1/users/profile
Authorization: Bearer <JWT_TOKEN>
```

### Venue Management

#### Create Venue (ORGANIZER/ADMIN)
```
POST /api/v1/venues
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "name": "Main Conference Hall",
  "address": "123 Main St, City",
  "capacity": 500,
  "description": "Large conference venue"
}
```

#### Get All Venues (PUBLIC)
```
GET /api/v1/venues
```

#### Get Venue by ID (PUBLIC)
```
GET /api/v1/venues/{id}
```

#### Update Venue (ORGANIZER/ADMIN)
```
PUT /api/v1/venues/{id}
Authorization: Bearer <JWT_TOKEN>
```

#### Delete Venue (ADMIN)
```
DELETE /api/v1/venues/{id}
Authorization: Bearer <JWT_TOKEN>
```

### Event Management

#### Create Event (ORGANIZER/ADMIN)
```
POST /api/v1/events
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "venue_id": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Tech Conference 2024",
  "description": "Annual technology conference",
  "event_date": "2024-12-01T10:00:00Z",
  "ticket_price": 99.99,
  "total_tickets": 200
}
```

#### Get All Events (PUBLIC)
```
GET /api/v1/events
```

#### Get Event by ID (PUBLIC)
```
GET /api/v1/events/{id}
```

#### Get My Events (ORGANIZER)
```
GET /api/v1/events/my-events
Authorization: Bearer <JWT_TOKEN>
```

#### Update Event (ORGANIZER/ADMIN)
```
PUT /api/v1/events/{id}
Authorization: Bearer <JWT_TOKEN>
```

#### Cancel Event (ORGANIZER/ADMIN)
```
PATCH /api/v1/events/{id}/cancel
Authorization: Bearer <JWT_TOKEN>
```

#### Delete Event (ORGANIZER/ADMIN)
```
DELETE /api/v1/events/{id}
Authorization: Bearer <JWT_TOKEN>
```

### Order Management

#### Create Order (USER)
```
POST /api/v1/orders
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "event_id": "123e4567-e89b-12d3-a456-426614174000",
  "quantity": 2
}
```

#### Get Order by ID (USER)
```
GET /api/v1/orders/{id}
Authorization: Bearer <JWT_TOKEN>
```

#### Get My Orders (USER)
```
GET /api/v1/orders/my-orders
Authorization: Bearer <JWT_TOKEN>
```

### Role-Based Access Control

- **PUBLIC**: Anyone can access
- **USER**: Authenticated users (can create orders)
- **ORGANIZER**: Can create/manage events and venues
- **ADMIN**: Full access to all operations

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

**Venues Table:**
```sql
CREATE TABLE venues (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    capacity INTEGER NOT NULL CHECK (capacity > 0),
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

**Events Table:**
```sql
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    venue_id UUID NOT NULL REFERENCES venues(id),
    organizer_id UUID NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    event_date TIMESTAMP NOT NULL,
    ticket_price DECIMAL(10,2) NOT NULL,
    total_tickets INTEGER NOT NULL,
    available_tickets INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'ACTIVE',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

**Orders Table:**
```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    event_id UUID NOT NULL REFERENCES events(id),
    quantity INTEGER NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Tickets Table:**
```sql
CREATE TABLE tickets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id),
    event_id UUID NOT NULL REFERENCES events(id),
    user_id UUID NOT NULL REFERENCES users(id),
    seat_info VARCHAR(50),
    qr_code VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
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

### Example API Workflow

#### 1. Create a User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "organizer@example.com",
    "username": "event_organizer",
    "password": "password123"
  }'
```

#### 2. Login to Get Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "organizer@example.com",
    "password": "password123"
  }'
```

#### 3. Create a Venue (ORGANIZER)
```bash
curl -X POST http://localhost:8080/api/v1/venues \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tech Conference Center",
    "address": "123 Innovation Drive",
    "capacity": 500,
    "description": "Modern conference facility"
  }'
```

#### 4. Create an Event (ORGANIZER)
```bash
curl -X POST http://localhost:8080/api/v1/events \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "venue_id": "<VENUE_ID>",
    "title": "Go Developer Conference",
    "description": "Annual Go programming conference",
    "event_date": "2024-12-01T10:00:00Z",
    "ticket_price": 99.99,
    "total_tickets": 200
  }'
```

#### 5. Create an Order (USER)
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer <USER_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "<EVENT_ID>",
    "quantity": 2
  }'
```

#### 6. View Browse Events (PUBLIC)
```bash
curl -X GET http://localhost:8080/api/v1/events
```

#### 7. View My Orders (USER)
```bash
curl -X GET http://localhost:8080/api/v1/orders/my-orders \
  -H "Authorization: Bearer <USER_TOKEN>"
```

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
   - Domain entities (User, Event, Venue, Order, Ticket)
   - Repository interfaces
   - Custom error types

2. **Infrastructure Layer** (`internal/infrastructure/`):
   - Database implementations (GORM)
   - JWT authentication service
   - Repository implementations
   - External service integrations

3. **Presentation Layer** (`internal/presentation/`):
   - HTTP handlers for all entities
   - Request/response mapping
   - Route registration
   - Middleware integration

4. **DTO Layer** (`internal/dto/`):
   - Data transfer objects
   - Request/response structures
   - JSON validation rules
   - Swagger annotations

### Event Ticketing System Flow

1. **Venue Creation**: ORGANIZER creates venues
2. **Event Creation**: ORGANIZER creates events at venues
3. **Event Browsing**: PUBLIC can view events and venues
4. **User Registration**: Anyone can register as USER
5. **Order Creation**: USER purchases tickets (atomic transaction)
6. **Ticket Generation**: System generates tickets for completed orders
7. **Order Management**: USER can view their orders

### Adding New Features

1. Define domain interfaces in `internal/domain/`
2. Create DTOs in `internal/dto/`
3. Implement infrastructure in `internal/infrastructure/`
4. Add HTTP handlers in `internal/presentation/`
5. Write comprehensive tests for all layers
6. Update Swagger documentation
7. Add database migrations if needed

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