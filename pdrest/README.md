# PD REST API

REST API server with PostgreSQL database support.

## Setup

### 1. Install Dependencies

```bash
go mod tidy
```

This will download the required dependencies including:
- `github.com/labstack/echo/v4` - Web framework
- `github.com/jackc/pgx/v5` - PostgreSQL driver

### 2. Configure Database

Create a `.env` file in the `pdrest` directory:

```env
DB_HOST=localhost
DB_PORT=5433
DB_USER=pdrest_user
DB_PASSWORD=1qaz2wsx
DB_NAME=pdrest
DB_SSLMODE=disable
```

### 3. Create Database

Make sure PostgreSQL is running and create the database:

```sql
CREATE DATABASE pdrest;
```

### 4. Run the Server

```bash
# Development
go run ./cmd/server

# Build
go build -o build/server.exe ./cmd/server

# Run built executable
./build/server.exe
```

## Features

- **PostgreSQL Integration**: Automatic connection pooling with pgx/v5
- **Fallback Support**: Falls back to in-memory repository if database is unavailable
- **Auto Schema**: Automatically creates database tables on startup
- **Connection Pooling**: Configurable connection pool with health checks

## Environment Variables

- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name
- `DB_SSLMODE` - SSL mode (default: disable)
- `DB_MAX_CONNS` - Maximum connections in pool (default: 25)
- `SERVER_PORT` - Server port (default: 8080)
- `SEED_DATA` - Set to "true" to seed initial data (optional)

## API Endpoints

- `GET /api/status` - Health check
- `GET /api/client_status/:id` - Get client status by ID

