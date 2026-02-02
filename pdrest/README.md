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
- `JWT_SECRET_KEY` - Secret key for JWT token signing (default: "your-secret-key-change-in-production")
- `JWT_ACCESS_TOKEN_TTL_HOURS` - Access token TTL in hours (default: 72)
- `JWT_REFRESH_TOKEN_TTL_HOURS` - Refresh token TTL in hours (default: 72)
- `JWT_STRICT_MODE` - If false, only checks that token is non-empty (default: true)
- `TELEGRAM_BOT_TOKEN` - Telegram bot token for hash verification (optional, if not set hash verification is disabled)

## API Documentation

### Text Documentation
- `GET /api/docs` - Get API documentation in text format

### Swagger/OpenAPI Documentation
- `GET /api/docs/openapi.yaml` - Get OpenAPI specification in YAML format
- `GET /api/docs/openapi.json` - Get OpenAPI specification in JSON format
- `GET /api/swagger/` - Interactive Swagger UI (browser-based documentation)

You can use these OpenAPI specs with Swagger UI or other OpenAPI tools:
- Visit `/api/swagger/` in your browser for interactive API documentation
- Import the YAML/JSON file into [Swagger Editor](https://editor.swagger.io/)
- Use with any OpenAPI-compatible tool

## API Endpoints

- `GET /api/status` - Health check
- `POST /api/auth/refresh` - Refresh JWT token (requires refresh_token in body)
- `GET /api/auth/status` - Check JWT authorization status, returns UUID if valid (requires JWT Bearer token)
- `GET /api/auth/google/verify` - Verify Google OAuth token and return JWT token pair (requires Google Bearer token in Authorization header)
- `GET /api/auth/telegram/verify` - Verify Telegram Web Login data and return JWT token pair (accepts Telegram auth data as query parameters or JSON body)
- `GET /api/user/last_login/:uuid` - Get user last login time by UUID (requires JWT Bearer token)
- `GET /api/user/profile/:uuid` - Get user profile (uuid and username) by UUID (requires JWT Bearer token)
- `POST /api/user/openbet` - Create a new bet, returns bet ID (requires JWT Bearer token, body contains side, sum, pair, timeframe, openPrice, openTime)
- `GET /api/user/betstatus?id=<bet_id>` - Get bet status with current price if timeframe has passed (requires JWT Bearer token)

