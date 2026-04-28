# Event Reservation API

RESTful backend service for the Event Reservation System, built with **Go**, **PostgreSQL**, and **RabbitMQ**. Uses hexagonal (ports & adapters) architecture for clean separation of concerns.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Project Structure](#project-structure)
- [Installation & Setup](#installation--setup)
- [Configuration](#configuration)
- [Running Locally](#running-locally)
- [API Endpoints](#api-endpoints)
- [Building & Deployment](#building--deployment)
- [Testing](#testing)
- [Database Migrations](#database-migrations)
- [Troubleshooting](#troubleshooting)

## Overview

This is the core API service that handles:

- **Event Management**: Create and manage events
- **Reservations**: Book, list, and manage reservations
- **Check-In**: Track attendance with check-in functionality
- **State Management**: Enforce reservation state transitions

### Architecture

The backend follows **hexagonal architecture** (ports & adapters):

```
pkg/
├── core/           # Domain logic (business rules, entities)
│   ├── domain/     # Core business objects (Reservation, Event)
│   └── services/   # Business logic (CheckIn, validation)
├── adapters/       # External integrations
│   ├── handlers/   # HTTP request handlers
│   ├── repositories/  # Database persistence (Postgres, DynamoDB)
│   └── messaging/  # Message broker integration (RabbitMQ)
├── bootstrap/      # Application initialization and wiring
└── platform/       # Platform-specific utilities
```

This architecture ensures:

- **Testability**: Core business logic is framework-independent
- **Flexibility**: Easy to swap adapters (e.g., different databases)
- **Maintainability**: Clear separation of concerns

## Prerequisites

- **Go**: 1.23 or later
- **PostgreSQL**: 13+ (or use Docker Compose)
- **Docker** & **Docker Compose**: For running PostgreSQL, RabbitMQ, LocalStack locally
- **Git**: For cloning the repository

### Verify Installation

```bash
go version          # Should be 1.23+
psql --version      # Should be psql 13+
docker --version
docker-compose --version
```

## Project Structure

```
backend/
├── cmd/
│   ├── api/        # HTTP API server entry point
│   │   └── main.go
│   └── worker/     # Background job worker (future use)
├── pkg/
│   ├── core/
│   │   ├── domain/     # Business entities
│   │   │   ├── reservation.go
│   │   │   └── event.go
│   │   └── services/   # Business logic
│   │       └── reservation.go
│   ├── adapters/
│   │   ├── handlers/   # HTTP handlers
│   │   │   └── reservation.go
│   │   ├── repositories/  # Data persistence
│   │   │   ├── postgres.go
│   │   │   └── in_memory.go
│   │   └── messaging/  # Event publishing
│   ├── bootstrap/      # Dependency injection & setup
│   │   └── server.go
│   └── platform/       # Utilities
├── migrations/
│   └── 001_initial_schema.sql
├── .env.example    # Example environment variables
├── go.mod          # Module definition
├── go.sum          # Dependency checksums
├── Dockerfile      # Container image definition
└── README.md       # This file
```

## Installation & Setup

### 1. Clone the Repository

```bash
cd /Users/femisowemimo/Documents/GitHub/booking-appointment
```

### 2. Install Dependencies

```bash
cd backend
go mod download
go mod verify
```

### 3. Set Up Environment Variables

```bash
cp .env.example .env
```

Then edit `.env` with your local configuration (see [Configuration](#configuration) section).

### 4. Start Infrastructure (Docker Compose)

```bash
cd ..  # Back to project root
docker-compose up -d
```

This starts:

- **PostgreSQL** on `localhost:5432`
- **RabbitMQ** on `localhost:5672` (management UI: http://localhost:15672)
- **LocalStack** on `localhost:4566`

Verify services are healthy:

```bash
docker-compose ps
```

### 5. Run Database Migrations

```bash
cd backend
psql -U user -h localhost -d appointments -f migrations/001_initial_schema.sql
```

Connection string: `postgres://user:password@localhost:5432/appointments?sslmode=disable`

## Configuration

### Environment Variables

Create a `.env` file in the `backend/` directory:

```bash
# Server Configuration
PORT=8080                    # HTTP server port (default: 8080)

# Database Configuration (PostgreSQL)
DATABASE_URL=postgres://user:password@localhost:5432/appointments?sslmode=disable

# Message Broker Configuration (RabbitMQ)
RABBITMQ_URL=amqp://user:password@localhost:5672/

# AWS Configuration (For LocalStack in development)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
AWS_ENDPOINT_URL=http://localhost:4566  # LocalStack endpoint
```

### Environment Profiles

| Profile        | DATABASE_URL       | RABBITMQ_URL   | Notes                           |
| -------------- | ------------------ | -------------- | ------------------------------- |
| **Local Dev**  | localhost:5432     | localhost:5672 | Use Docker Compose services     |
| **Production** | Cloud RDS instance | Cloud RabbitMQ | Managed services on Railway/AWS |

## Running Locally

### Option 1: Direct Go Execution

```bash
cd backend
go run cmd/api/main.go
```

Output:

```
Starting Reservation API Service...
HTTP Server listening on 127.0.0.1:8080
```

### Option 2: Build & Run Binary

```bash
cd backend

# Build
go build -o api cmd/api/main.go

# Run
./api
```

### Option 3: Using the Provided Script

```bash
cd /Users/femisowemimo/Documents/GitHub/booking-appointment
./start-backend.sh
```

### Verify API is Running

```bash
curl http://localhost:8080/api/reservations
```

Expected response (HTTP 200):

```json
[]
```

## API Endpoints

### Get All Reservations

```http
GET /api/reservations
```

**Response:**

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "eventId": "123",
    "guestEmail": "john@example.com",
    "numberOfGuests": 2,
    "status": "CONFIRMED",
    "createdAt": "2026-04-28T10:00:00Z",
    "version": 1
  }
]
```

### Create Reservation

```http
POST /api/reservations
Content-Type: application/json

{
  "eventId": "123",
  "guestEmail": "john@example.com",
  "numberOfGuests": 2
}
```

**Response (201 Created):**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "eventId": "123",
  "guestEmail": "john@example.com",
  "numberOfGuests": 2,
  "status": "CONFIRMED",
  "createdAt": "2026-04-28T10:00:00Z",
  "version": 1
}
```

### Check In Reservation

```http
POST /api/reservations/{reservationId}/checkin
```

**Response (200 OK):**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "eventId": "123",
  "guestEmail": "john@example.com",
  "numberOfGuests": 2,
  "status": "CHECKED_IN",
  "createdAt": "2026-04-28T10:00:00Z",
  "version": 2
}
```

**Possible Errors:**

- `404 Not Found`: Reservation doesn't exist
- `409 Conflict`: Reservation status doesn't allow check-in (e.g., already cancelled)

### State Transitions

Reservations follow a state machine:

```
CONFIRMED → CHECKED_IN
          ↘ CANCELLED (anytime)

CHECKED_IN → COMPLETED
           ↘ CANCELLED (only if policy allows)
```

Check-in is only allowed from `CONFIRMED` state.

## Building & Deployment

### Build Release Binary

```bash
cd backend

# Default (local binary)
go build -o api cmd/api/main.go

# With version info embedded (recommended for production)
VERSION=$(git describe --tags --always)
go build -ldflags="-X main.Version=$VERSION" -o api cmd/api/main.go
```

### Build Docker Image

```bash
cd ..  # Project root
docker build -f backend/Dockerfile -t booking-appointment-api:latest .
```

Run the container:

```bash
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://user:password@host:5432/appointments?sslmode=disable" \
  -e RABBITMQ_URL="amqp://user:password@host:5672/" \
  booking-appointment-api:latest
```

### Deploy to Production (Railway)

The project uses Railway for production deployment. Configuration:

1. Connect your GitHub repo to Railway
2. Railway automatically detects `backend/Dockerfile`
3. Set environment variables in Railway dashboard:
   - `DATABASE_URL` (Railway Postgres)
   - `RABBITMQ_URL` (Railway RabbitMQ)
   - `PORT` (Railway assigns dynamically)

Check [RAILWAY.md](../RAILWAY.md) for detailed instructions.

## Testing

### Run All Tests

```bash
cd backend
go test ./...
```

### Run Tests with Verbose Output

```bash
go test -v ./...
```

### Run Tests with Coverage

```bash
go test -cover ./...
```

### Test Specific Package

```bash
go test ./pkg/core/services/...
```

### Integration Tests (Requires Running Services)

```bash
# Start services first
docker-compose up -d

# Run integration tests
go test -tags=integration ./...
```

## Database Migrations

### View Migration History

```bash
psql -U user -h localhost -d appointments -c "\dt _prisma_migrations"
```

### Create New Migration

Create a SQL file in `migrations/`:

```bash
touch migrations/002_add_feature.sql
```

Write your migration:

```sql
BEGIN;

-- Your DDL statements here
ALTER TABLE reservations ADD COLUMN new_column VARCHAR(255);

COMMIT;
```

Apply it:

```bash
psql -U user -h localhost -d appointments -f migrations/002_add_feature.sql
```

## Troubleshooting

### Issue: "Connection refused" on `localhost:8080`

**Solution:**

1. Check if API is running: `ps aux | grep "api"`
2. Check if port 8080 is in use: `lsof -i :8080`
3. Verify `.env` is in the `backend/` directory

### Issue: Database connection error

**Error:** `could not connect to postgres server`

**Solutions:**

1. Verify PostgreSQL is running: `docker-compose ps postgres`
2. Check `DATABASE_URL` in `.env` matches your setup
3. Verify credentials: `psql -U user -h localhost -d appointments`
4. Check migrations were applied: `psql -U user -h localhost -d appointments -c "\dt"`

### Issue: RabbitMQ connection error

**Error:** `connection refused on localhost:5672`

**Solutions:**

1. Start RabbitMQ: `docker-compose up -d rabbitmq`
2. Verify it's running: `docker-compose ps rabbitmq`
3. Check `RABBITMQ_URL` in `.env`

### Issue: "Address already in use" on port 8080

**Solution:** Use a different port or kill the existing process:

```bash
# Find process on port 8080
lsof -i :8080

# Kill it
kill -9 <PID>

# Or use a different port
PORT=9000 go run cmd/api/main.go
```

### Issue: Tests fail with "database error"

**Solution:**

1. Ensure Docker Compose services are running: `docker-compose up -d`
2. Check database connectivity: `psql -U user -h localhost -d appointments`
3. Run migrations: `psql -U user -h localhost -d appointments -f migrations/001_initial_schema.sql`

### Issue: API returns 404 on `/api/reservations/...` endpoint

**Solution:**

- Ensure the backend is running on `localhost:8080`
- From mobile/web, verify requests are being proxied correctly
- Check backend logs: `go run cmd/api/main.go` shows all HTTP requests

## Development Workflow

### Local Development Loop

1. **Start services:**

   ```bash
   docker-compose up -d
   ```

2. **Run migrations:**

   ```bash
   psql -U user -h localhost -d appointments -f backend/migrations/001_initial_schema.sql
   ```

3. **Start API:**

   ```bash
   cd backend && go run cmd/api/main.go
   ```

4. **Make code changes** → Go reloads automatically (with `air` tool, optional)

5. **Test changes:**
   ```bash
   curl -X POST http://localhost:8080/api/reservations \
     -H "Content-Type: application/json" \
     -d '{"eventId":"123","guestEmail":"test@example.com","numberOfGuests":2}'
   ```

### Optional: Use `air` for Auto-Reload

Install:

```bash
go install github.com/cosmtrek/air@latest
```

Create `.air.toml` in the `backend/` directory and run:

```bash
air
```

This recompiles and runs the app whenever you save a Go file.

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [RabbitMQ Tutorials](https://www.rabbitmq.com/getstarted.html)
- [Railway Documentation](https://docs.railway.app/)
- [Docker Documentation](https://docs.docker.com/)

## Support

For issues or questions:

1. Check [Troubleshooting](#troubleshooting) section
2. Review backend logs: `go run cmd/api/main.go`
3. Verify `.env` configuration
4. Check Docker Compose services: `docker-compose ps`
