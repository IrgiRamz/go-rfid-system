# Go API for Milk Run RFID System

Laravel-compatible REST API built with Go (Gin framework) using the same database as the Laravel application.

## Requirements

- Go 1.21+
- Access to the same MySQL database as Laravel
- Laravel app already running (or at least the database)

## Quick Start

### 1. Install Dependencies

```bash
cd app_golang
go mod tidy
```

### 2. Create Environment File

```bash
cp .env.example .env
```

Edit `.env` with your database settings:

```env
APP_PORT=8080
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=rfid_app
```

### 3. Run the Server

```bash
go run main.go
```

Server will start on `http://localhost:8080`

## Project Structure

```
app_golang/
├── main.go                    # Entry point
├── go.mod                     # Go modules
├── .env.example               # Environment template
├── .gitignore                 # Git ignore
├── config/
│   └── database.go            # Database connection
├── models/
│   ├── user.go                # User model
│   └── personal_access_token.go # Sanctum token model
├── handlers/
│   ├── auth_handler.go        # Login & Logout
│   ├── transaction_handler.go # Transactions
│   ├── master_data_handler.go # Master data
│   └── health_handler.go      # Health check
├── middleware/
│   ├── sanctum_auth.go        # Sanctum authentication
│   └── rate_limit.go          # Rate limiting
├── routes/
│   └── routes.go              # Route definitions
├── helpers/
│   └── response.go            # JSON response helpers
└── docs/
    ├── README.md              # This file
    ├── API_MAPPING.md         # Endpoint mapping
    └── SANCTUM_COMPATIBILITY.md # Token compatibility
```

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/health` | No | Health check |
| POST | `/api/login` | No | User login |
| POST | `/api/logout` | Yes | User logout |
| GET | `/api/master-data` | Yes | Get suppliers & routes |
| GET | `/api/master-rfid` | Yes | Get RFID tags |
| GET | `/api/transactions/history` | Yes | Get transaction history |
| POST | `/api/transactions/submit` | Yes | Submit new transaction |

## Testing with cURL

### Health Check

```bash
curl http://localhost:8080/api/health
```

Response:
```json
{"success":true,"message":"Go API is running"}
```

### Login

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "login_id": "admin",
    "password": "admin123"
  }'
```

Response:
```json
{
    "status": "success",
    "message": "Login berhasil.",
    "data": {
        "user": {
            "id": 1,
            "name": "Administrator",
            "username": "admin",
            "email": "admin@gmail.com",
            "role": "admin"
        },
        "token": "1|abc..."
    }
}
```

### Using Protected Endpoints

```bash
# Save the token from login response
TOKEN="1|abc..."

# Get Master Data
curl -X GET http://localhost:8080/api/master-data \
  -H "Authorization: Bearer $TOKEN"

# Get Transaction History
curl -X GET http://localhost:8080/api/transactions/history \
  -H "Authorization: Bearer $TOKEN"

# Submit Transaction
curl -X POST http://localhost:8080/api/transactions/submit \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_date": "24-06-2026 13:22",
    "supplier_code": "7600",
    "lp_name": "ARYOS",
    "route_code": "RWC-09D",
    "status_truck": "IN FROM WJ",
    "no_truck_epc": "AF25110000000003",
    "scanned_items": [
      {"epc_code": "BB25110000000006", "type": "Pallet"},
      {"epc_code": "BB25110000000131", "type": "Box"}
    ]
  }'

# Logout
curl -X POST http://localhost:8080/api/logout \
  -H "Authorization: Bearer $TOKEN"
```

## Testing with Postman

1. **Create a new Collection** for RFID API
2. **Add Environment** with variables:
   - `base_url`: `http://localhost:8080/api`
   - `token`: (empty, will be set from login response)
3. **Add Request: Login**
   - Method: POST
   - URL: `{{base_url}}/login`
   - Body (JSON):
     ```json
     {
       "login_id": "admin",
       "password": "admin123"
     }
     ```
   - Tests:
     ```javascript
     pm.test("Login successful", function() {
         var jsonData = pm.response.json();
         pm.expect(jsonData.status).to.eql("success");
         pm.environment.set("token", jsonData.data.token);
     });
     ```
4. **Add Request: Get Master Data**
   - Method: GET
   - URL: `{{base_url}}/master-data`
   - Headers: `Authorization: Bearer {{token}}`
5. **Add Request: Submit Transaction**
   - Method: POST
   - URL: `{{base_url}}/transactions/submit`
   - Headers: `Authorization: Bearer {{token}}`
   - Body (JSON): See example above

## Cross-Compatibility Testing

### Using Laravel Token in Go API

1. Login via Laravel (port 8000):
   ```bash
   curl -X POST http://localhost:8000/api/login \
     -H "Content-Type: application/json" \
     -d '{"login_id": "admin", "password": "admin123"}'
   ```

2. Use the token from Laravel in Go API (port 8080):
   ```bash
   curl -X GET http://localhost:8080/api/master-data \
     -H "Authorization: Bearer {laravel_token}"
   ```

### Using Go Token in Laravel API

1. Login via Go (port 8080):
   ```bash
   curl -X POST http://localhost:8080/api/login \
     -H "Content-Type: application/json" \
     -d '{"login_id": "admin", "password": "admin123"}'
   ```

2. Use the token from Go in Laravel API (port 8000):
   ```bash
   curl -X GET http://localhost:8000/api/master-data \
     -H "Authorization: Bearer {go_token}"
   ```

Both should work seamlessly because they use the same token format and validation method.

## Database

The Go API uses the **same database** as the Laravel application. No migrations are needed.

Tables used:
- `users` - User accounts
- `personal_access_tokens` - Sanctum tokens
- `master_suppliers` - Supplier reference data
- `master_routes` - Route reference data
- `master_rfid_tags` - RFID tag reference data
- `milk_run_logs` - Transaction logs
- `scanned_items` - Scanned items per transaction

## Rate Limiting

| Endpoint | Limit |
|----------|-------|
| `/api/login` | 7 requests per minute per IP |
| Protected endpoints | 200 requests per minute per IP |

## Error Responses

All errors follow this format:

```json
{
    "status": "error",
    "message": "Error description"
}
```

HTTP Status Codes:
- 200: Success
- 201: Created
- 401: Unauthorized
- 403: Forbidden (account inactive)
- 422: Validation Error
- 429: Too Many Requests
- 500: Internal Server Error

## Security

- Passwords are verified using bcrypt
- Tokens use SHA-256 hashing (compatible with Laravel Sanctum)
- Rate limiting prevents brute force attacks
- No sensitive data exposed in error messages
- Database credentials loaded from environment variables
