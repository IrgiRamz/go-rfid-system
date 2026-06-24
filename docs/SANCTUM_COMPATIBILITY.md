# Sanctum Token Compatibility

This document explains how Laravel Sanctum tokens work and how the Go API implements full compatibility.

## Token Format

Laravel Sanctum generates tokens in the format:

```
{token_id}|{plain_text_token}
```

Example:
```
1|aBcDeFgHiJkLmNoPqRsTuVwXyZ123456
```

- **token_id**: Auto-increment integer (from `personal_access_tokens.id`)
- **plain_text_token**: 40-character random string (base64url)

## Database Schema

### personal_access_tokens Table

```sql
CREATE TABLE personal_access_tokens (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tokenable_type VARCHAR(255)     -- e.g., "App\\Models\\User"
    tokenable_id BIGINT UNSIGNED    -- e.g., 1
    name TEXT
    token VARCHAR(64) UNIQUE        -- SHA-256 hash of plain_text_token
    abilities TEXT NULL
    last_used_at TIMESTAMP NULL
    expires_at TIMESTAMP NULL
    created_at TIMESTAMP
    updated_at TIMESTAMP
);
```

### Token Hashing

Laravel Sanctum uses SHA-256 to hash the plain text token:

```go
// Go implementation
import "crypto/sha256"
import "encoding/hex"

func hashToken(plainToken string) string {
    hash := sha256.Sum256([]byte(plainToken))
    return hex.EncodeToString(hash[:])
}
```

### Token Validation Flow

1. Mobile app sends: `Authorization: Bearer {id}|{plain_text_token}`
2. Server splits by `|`: `{id}` and `{plain_text_token}`
3. Hash the plain text token with SHA-256
4. Query: `SELECT * FROM personal_access_tokens WHERE id = {id} AND token = {hash}`
5. Check `expires_at` if present
6. Update `last_used_at` to current timestamp
7. Load associated user via `tokenable_id` and `tokenable_type`
8. Allow request to proceed

## Cross-Compatibility

### Using Laravel Token in Go API

1. User logs in via Laravel API
2. Gets token: `1|aBcDeFgHiJkLmNoPqRsTuVwXyZ123456`
3. Uses this token in Go API: `Authorization: Bearer 1|aBcDeFgHiJkLmNoPqRsTuVwXyZ123456`
4. Go API validates by:
   - Extracting ID = `1`
   - Extracting plain token = `aBcDeFgHiJkLmNoPqRsTuVwXyZ123456`
   - Hashing: `sha256("aBcDeFgHiJkLmNoPqRsTuVwXyZ123456")`
   - Looking up in `personal_access_tokens` where `id = 1` AND `token = {hash}`
5. ✅ Token is valid, request proceeds

### Using Go Token in Laravel API

1. User logs in via Go API
2. Gets token: `1|GoGeneratedToken12345678901234567890`
3. Uses this token in Laravel API: `Authorization: Bearer 1|GoGeneratedToken12345678901234567890`
4. Laravel Sanctum validates by:
   - Extracting ID = `1`
   - Extracting plain token = `GoGeneratedToken12345678901234567890`
   - Hashing: `hash("GoGeneratedToken12345678901234567890")` (SHA-256)
   - Looking up in `personal_access_tokens` where `id = 1` AND `token = {hash}`
5. ✅ Token is valid, request proceeds

## Token Generation (Go)

```go
import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "encoding/base64"
)

func generatePlainToken() (string, error) {
    bytes := make([]byte, 40)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    // Use base64url encoding (URL-safe, no padding)
    return base64.URLEncoding.EncodeToString(bytes), nil
}

func hashToken(plainToken string) string {
    hash := sha256.Sum256([]byte(plainToken))
    return hex.EncodeToString(hash[:])
}
```

## Abilities

Laravel Sanctum tokens can have abilities (permissions). In this project:

- Login creates token with ability: `['*']` (full access)
- Abilities are stored as JSON: `'["*"]'`
- Go API does not enforce abilities, only validates token existence

## Expiration

- By default, Laravel Sanctum tokens do not expire
- The `expires_at` column is optional
- Go API checks `expires_at` if it's not null
- If expired, token is rejected with 401

## Testing Cross-Compatibility

### Test 1: Laravel Token → Go API

```bash
# 1. Login via Laravel (port 8000)
curl -X POST http://localhost:8000/api/login \
  -H "Content-Type: application/json" \
  -d '{"login_id": "admin", "password": "admin123"}'

# Response:
# {"status": "success", "message": "Login berhasil.", "data": {"user": {...}, "token": "2|abc..."}}

# 2. Use Laravel token in Go API (port 8080)
curl -X GET http://localhost:8080/api/master-data \
  -H "Authorization: Bearer 2|abc..."

# Expected: 200 OK with master data
```

### Test 2: Go Token → Laravel API

```bash
# 1. Login via Go (port 8080)
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"login_id": "admin", "password": "admin123"}'

# Response:
# {"status": "success", "message": "Login berhasil.", "data": {"user": {...}, "token": "3|xyz..."}}

# 2. Use Go token in Laravel API (port 8000)
curl -X GET http://localhost:8000/api/master-data \
  -H "Authorization: Bearer 3|xyz..."

# Expected: 200 OK with master data
```

### Test 3: Token Revocation

```bash
# 1. Login via Go
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"login_id": "admin", "password": "admin123"}'

# 2. Logout via Go
curl -X POST http://localhost:8080/api/logout \
  -H "Authorization: Bearer {token}"

# 3. Try to use same token in Laravel
curl -X GET http://localhost:8000/api/master-data \
  -H "Authorization: Bearer {token}"

# Expected: 401 Unauthorized (token was deleted)
```

## Security Notes

1. **Plain token is only returned once** at login time
2. **Only the hash is stored** in the database
3. **Token revocation** deletes the record from `personal_access_tokens`
4. **Rate limiting** prevents brute force attacks on login
5. **Token is tied to user ID** preventing token hijacking

## Implementation Files

- `middleware/sanctum_auth.go` - Token validation middleware
- `models/personal_access_token.go` - Token model
- `handlers/auth_handler.go` - Login (token creation) and Logout (token deletion)