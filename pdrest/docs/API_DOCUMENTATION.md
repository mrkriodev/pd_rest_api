# PD REST API Documentation

## Base URL
All API endpoints are prefixed with `/api`

## Authentication

Most endpoints require JWT Bearer token authentication. Include the token in the Authorization header:
```
Authorization: Bearer <your_jwt_token>
```

---

## Endpoints

### Health Check

#### GET /api/status
Health check endpoint.

**Response:**
```json
{
  "status": "ok"
}
```

---

### Events

#### GET /api/available_events
Get list of available events.

**Response:**
```json
{
  "events": [
    {
      "id": "event_id",
      "badge": "Event badge",
      "title": "Event title",
      "desc": "Event description",
      "deadline": "2025-12-01T12:00:00Z",
      "tags": "global",
      "reward": [...],
      "info": "Additional info"
    }
  ]
}
```

---

### Authentication

#### POST /api/auth/refresh
Refresh JWT token using refresh token.

**Request Body:**
```json
{
  "refresh_token": "your_refresh_token"
}
```

**Response:**
```json
{
  "access_token": "new_access_token",
  "refresh_token": "new_refresh_token",
  "expires_in": 3600
}
```

#### GET /api/auth/status
Check JWT authorization status. Returns user UUID if token is valid.

**Headers:**
- `Authorization: Bearer <jwt_token>` (required)

**Response:**
```json
{
  "uuid": "user-uuid-here"
}
```

**Error Response (401):**
```json
{
  "error": "unauthorized"
}
```

#### GET /api/auth/google/verify
Verify Google OAuth token and return JWT token pair.

**Headers:**
- `Authorization: Bearer <google_token>` (required)

**Response:**
```json
{
  "access_token": "jwt_access_token",
  "refresh_token": "jwt_refresh_token",
  "expires_in": 3600
}
```

**Error Response (401):**
```json
{
  "error": "invalid token"
}
```

#### POST /api/auth/google/registeroauth2
Google OAuth2 code registration (registers user) and returns JWT token pair.

**Headers:**
- `X-SESSION-ID` (required)

**Request Body:**
```json
{
  "code": "<oauth2_code>",
  "redirect_uri": "<optional_redirect_uri>"
}
```

**Response:**
```json
{
  "userID": "user-uuid",
  "access_token": "jwt_access_token",
  "refresh_token": "jwt_refresh_token",
  "expires_in": 3600
}
```

**Error Response (401):**
```json
{
  "error": "failed to exchange code"
}
```

#### GET /api/auth/google/callback
Google OAuth2 callback (returns JWT).

**Query Parameters:**
- `code` - OAuth2 authorization code

**Headers (optional):**
- `code` - OAuth2 authorization code (if not provided in query)

**Response:**
```json
{
  "userID": "user-uuid",
  "access_token": "jwt_access_token",
  "refresh_token": "jwt_refresh_token",
  "expires_in": 3600
}
```

**Error Response (401):**
```json
{
  "error": "failed to exchange code"
}
```

#### GET /api/googlecallback
Google OAuth2 callback alias (returns JWT).

**Query Parameters:**
- `code` - OAuth2 authorization code

**Response:**
```json
{
  "userID": "user-uuid",
  "access_token": "jwt_access_token",
  "refresh_token": "jwt_refresh_token",
  "expires_in": 3600
}
```

**Error Response (401):**
```json
{
  "error": "failed to exchange code"
}
```

#### POST /api/auth/telegram/login
Telegram login (registers user) and returns JWT token pair.

**Query Parameters (or JSON body):**
- `id` - Telegram user ID
- `first_name` - User first name
- `last_name` - User last name (optional)
- `username` - Telegram username (optional)
- `photo_url` - User photo URL (optional)
- `auth_date` - Authentication timestamp
- `hash` - Verification hash

**Response:**
```json
{
  "access_token": "jwt_access_token",
  "refresh_token": "jwt_refresh_token",
  "expires_in": 3600
}
```

**Error Response (401):**
```json
{
  "error": "invalid hash"
}
```

#### GET /api/auth/telegram/callback
Telegram WebApp callback (returns JWT for existing user).

**Query Parameters:**
- `tgWebAppData` - `window.Telegram.WebApp.tgInitData`

**Response:**
```json
{
  "userID": "user-uuid",
  "access_token": "jwt_access_token",
  "refresh_token": "jwt_refresh_token",
  "expires_in": 3600
}
```

**Error Response (401):**
```json
{
  "error": "invalid hash"
}
```
**Error Response (404):**
```json
{
  "error": "user not found"
}
```

#### POST /api/auth/telegram/webapp
Telegram WebApp login (registers user) and returns JWT token pair.

**Request Body:**
```json
{
  "tgInitData": "<window.Telegram.WebApp.tgInitData>"
}
```

**Response:**
```json
{
  "userID": "user-uuid",
  "access_token": "jwt_access_token",
  "refresh_token": "jwt_refresh_token",
  "expires_in": 3600
}
```

**Error Response (401):**
```json
{
  "error": "invalid hash"
}
```

---

### User Endpoints (Protected by JWT)

All endpoints in this section require JWT Bearer token in Authorization header.

#### GET /api/user/last_login/:uuid
Get user last login time by UUID.

**Headers:**
- `Authorization: Bearer <jwt_token>` (required)

**Path Parameters:**
- `uuid` - User UUID

**Response:**
```json
{
  "uuid": "user-uuid",
  "last_login_at": 1733054400000
}
```

**Error Response (404):**
```json
{
  "error": "user not found"
}
```

#### GET /api/user/profile/:uuid
Get user profile (UUID and username) by UUID.

**Headers:**
- `Authorization: Bearer <jwt_token>` (required)

**Path Parameters:**
- `uuid` - User UUID

**Response:**
```json
{
  "uuid": "user-uuid",
  "username": "John Doe"
}
```

**Error Response (404):**
```json
{
  "error": "user not found"
}
```

#### POST /api/user/openbet
Create a new bet. Returns bet ID.

**Headers:**
- `Authorization: Bearer <jwt_token>` (required)

**Request Body:**
```json
{
  "side": "pump",
  "sum": 1000,
  "pair": "ETH/USDT",
  "timeframe": 60,
  "openPrice": 2765,
  "openTime": "2025-11-09T12:35:00Z"
}
```

**Request Fields:**
- `side` (string, required) - Bet side: "pump" or "dump"
- `sum` (number, required) - Bet amount (must be > 0)
- `pair` (string, required) - Trading pair (e.g., "ETH/USDT")
- `timeframe` (integer, required) - Timeframe in seconds (must be > 0)
- `openPrice` (number, required) - Opening price (must be > 0)
- `openTime` (string, required) - Opening time in ISO 8601 format

**Response:**
```json
{
  "id": 123
}
```

**Error Response (400):**
```json
{
  "error": "side must be 'pump' or 'dump'"
}
```

#### GET /api/user/betstatus?id=<bet_id>
Get bet status with current price if timeframe has passed.

**Headers:**
- `Authorization: Bearer <jwt_token>` (required)

**Query Parameters:**
- `id` (required) - Bet ID

**Response:**
```json
{
  "side": "pump",
  "sum": 1000,
  "pair": "ETH/USDT",
  "timeframe": 60,
  "openPrice": 2765,
  "closePrice": 2785,
  "openTime": "2025-11-09T12:35:00Z"
}
```

**Response Fields:**
- `side` - Bet side ("pump" or "dump")
- `sum` - Bet amount
- `pair` - Trading pair
- `timeframe` - Timeframe in seconds
- `openPrice` - Opening price
- `closePrice` - Closing price (null if timeframe hasn't passed yet)
- `openTime` - Opening time

**Note:** If the timeframe has passed and `closePrice` is not set, the system will automatically fetch the current price from Binance API and update the bet.

**Error Response (404):**
```json
{
  "error": "bet not found"
}
```

#### GET /api/getidbysession
Get user UUID by session_id + IP (derived preauth token).

**Headers:**
- `X-SESSION-ID` (required)
- `X-Forwarded-For` or `X-Real-IP` (optional)

**Response:**
```json
{
  "userId": "user-uuid"
}
```

#### GET /api/user/unfinished_bets/:uuid
Get unfinished bets (closePrice is null) for a user.

**Headers:**
- `Authorization: Bearer <jwt_token>` (required)

**Path Parameters:**
- `uuid` (required) - User UUID

**Response:**
```json
{
  "bets": [
    {
      "id": 123,
      "userID": "user-uuid",
      "side": "pump",
      "sum": 1000,
      "pair": "ETH/USDT",
      "timeframe": 60,
      "openPrice": 2765,
      "openTime": "2025-11-09T12:35:00Z",
      "closePrice": null,
      "closeTime": null
    }
  ]
}
```

---

### Roulette Endpoints

#### GET /api/roulette/status
Get roulette status by preauth token.

**Query Parameters or Headers:**
- `preauth_token` (required) - Preauth token (query parameter or X-Preauth-Token header)

**Response:**
```json
{
  "config": {...},
  "canSpin": true,
  "remainingSpins": 3,
  "prizeTaken": false,
  "roulette": {...}
}
```

#### POST /api/roulette/spin
Perform a spin using preauth token.

**Headers:**
- `Authorization: Bearer <token>` (required for roulette_id != 1; optional for roulette_id = 1; if provided, preauth_token is linked to this user)
- `X-Preauth-Token` (optional)
- `X-SESSION-ID` (optional, required if preauth_token is not provided)

**Request Body:**
```json
{
  "preauth_token": "token_here"
}
```

**Response:**
```json
{
  "roulette": {...},
  "remainingSpins": 2,
  "canSpin": true
}
```

#### POST /api/roulette/take-prize
Take prize after completing all spins.

**Headers:**
- `Authorization: Bearer <token>` (required for roulette_id != 1; optional for roulette_id = 1; if provided, preauth_token is linked to this user)
- `X-Preauth-Token` (optional)
- `X-SESSION-ID` (optional, required if preauth_token is not provided)

**Request Body:**
```json
{
  "preauth_token": "token_here"
}
```

**Response:**
```json
{
  "success": true,
  "prize": "Prize name",
  "message": "Prize taken successfully"
}
```

#### POST /api/roulette/preauth-token
Create a preauth token.

**Request Body:**
```json
{
  "token": "token_string",
  "type": "on_start",
  "event_id": "event_id_here",
  "expires_at": 1733054400000
}
```

**Response:**
```json
{
  "success": true,
  "message": "Preauth token created successfully"
}
```

---

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
```json
{
  "error": "error message"
}
```

### 401 Unauthorized
```json
{
  "error": "unauthorized"
}
```

### 404 Not Found
```json
{
  "error": "resource not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error"
}
```

---

## Authentication Flow

1. **Get JWT Token:**
   - Use `/api/auth/google/verify` with Google token, OR
   - Use `/api/auth/google/registeroauth2` with Google OAuth2 code, OR
   - Use `/api/auth/google/callback` with Google OAuth2 code (redirect), OR
   - Use `/api/googlecallback` with Google OAuth2 code (redirect alias), OR
   - Use `/api/auth/telegram/login` with Telegram auth data, OR
   - Use `/api/auth/telegram/callback` for Telegram web login redirect, OR
   - Use `/api/auth/telegram/webapp` with Telegram WebApp tgInitData

2. **Use JWT Token:**
   - Include token in `Authorization: Bearer <token>` header for protected endpoints

3. **Refresh Token:**
   - Use `/api/auth/refresh` with refresh_token when access token expires

---

## Examples

### Create a Bet
```bash
curl -X POST http://localhost:8080/api/user/openbet \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "side": "pump",
    "sum": 1000,
    "pair": "ETH/USDT",
    "timeframe": 60,
    "openPrice": 2765,
    "openTime": "2025-11-09T12:35:00Z"
  }'
```

### Get Bet Status
```bash
curl -X GET "http://localhost:8080/api/user/betstatus?id=123" \
  -H "Authorization: Bearer <jwt_token>"
```

### Verify Google Token
```bash
curl -X GET http://localhost:8080/api/auth/google/verify \
  -H "Authorization: Bearer <google_token>"
```

