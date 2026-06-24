# API Mapping: Laravel vs Go

## Endpoint Overview

| Laravel Endpoint | Go Endpoint | Method | Auth | Rate Limit | Status |
|------------------|-------------|--------|------|------------|--------|
| `/api/login` | `/api/login` | POST | No | 7/min | ✅ Implemented |
| `/api/logout` | `/api/logout` | POST | Yes | 200/min | ✅ Implemented |
| `/api/transactions/submit` | `/api/transactions/submit` | POST | Yes | 200/min | ✅ Implemented |
| `/api/transactions/history` | `/api/transactions/history` | GET | Yes | 200/min | ✅ Implemented |
| `/api/master-data` | `/api/master-data` | GET | Yes | 200/min | ✅ Implemented |
| `/api/master-rfid` | `/api/master-rfid` | GET | Yes | 200/min | ✅ Implemented |
| - | `/api/health` | GET | No | - | ✅ Implemented |

## Endpoint Details

### 1. POST /api/login

**Laravel:**
- File: `app/Http/Controllers/Api/AuthController.php`
- Method: `login()`
- Rate Limit: 7 requests per minute per IP

**Go:**
- File: `handlers/auth_handler.go`
- Function: `Login()`
- Rate Limit: 7 requests per minute per IP

**Request Body:**
```json
{
    "login_id": "admin@yimm.com",
    "password": "admin123"
}
```

**Success Response (200):**
```json
{
    "status": "success",
    "message": "Login berhasil.",
    "data": {
        "user": {
            "id": 1,
            "name": "Administrator",
            "username": "admin",
            "email": "admin@yimm.com",
            "role": "admin"
        },
        "token": "1|aBcDeFgH..."
    }
}
```

**Error Responses:**
- 401: `{"status": "error", "message": "Username/Email atau password salah."}`
- 403: `{"status": "error", "message": "Akun Anda tidak aktif. Hubungi administrator."}`

**Differences:** None

---

### 2. POST /api/logout

**Laravel:**
- File: `app/Http/Controllers/Api/AuthController.php`
- Method: `logout()`
- Rate Limit: 200 requests per minute per IP

**Go:**
- File: `handlers/auth_handler.go`
- Function: `Logout()`
- Rate Limit: 200 requests per minute per IP

**Request Body:** None (Bearer Token required)

**Success Response (200):**
```json
{
    "status": "success",
    "message": "Logout berhasil."
}
```

**Error Responses:**
- 401: `{"status": "error", "message": "Unauthenticated."}`

**Differences:** None

---

### 3. POST /api/transactions/submit

**Laravel:**
- File: `app/Http/Controllers/Api/TransactionController.php`
- Method: `submit()`
- Rate Limit: 200 requests per minute per IP

**Go:**
- File: `handlers/transaction_handler.go`
- Function: `Submit()`
- Rate Limit: 200 requests per minute per IP

**Request Body:**
```json
{
    "transaction_date": "18-12-2026 13:22",
    "supplier_code": "7600",
    "lp_name": "ARYOS",
    "route_code": "RWC-09D",
    "status_truck": "IN FROM WJ",
    "no_truck_epc": "AF25110000000003",
    "scanned_items": [
        {"epc_code": "BB25110000000006", "type": "Pallet"},
        {"epc_code": "BB25110000000131", "type": "Box"}
    ]
}
```

**Validation Rules:**
- `transaction_date`: required, format `d-m-Y H:i`
- `supplier_code`: required, must exist in `master_suppliers`
- `lp_name`: required
- `route_code`: required, must exist in `master_routes`
- `status_truck`: required, enum: `IN FROM WJ`, `IN FROM HO`, `OUT TO WJ`, `OUT TO HO`
- `no_truck_epc`: required, must exist in `master_rfid_tags` AND category must be `Truck`
- `scanned_items`: required, array, min 1 item
- `scanned_items.*.epc_code`: required, must exist in `master_rfid_tags`
- `scanned_items.*.type`: required, enum: `Pallet`, `Box`

**Success Response (201):**
```json
{
    "status": "success",
    "message": "Transaksi berhasil disimpan. Email sedang diproses.",
    "data": {
        "log_id": 42,
        "ritase": 1
    }
}
```

**Error Responses:**
- 401: `{"status": "error", "message": "Unauthenticated."}`
- 422: `{"status": "error", "message": "..."}` (custom validation errors)
- 422: `{"message": "The given data was invalid.", "errors": {...}}` (Laravel validation format)
- 500: `{"status": "error", "message": "Terjadi kesalahan saat menyimpan transaksi: ..."}`

**Differences:** None (includes database transaction for data integrity)

---

### 4. GET /api/transactions/history

**Laravel:**
- File: `app/Http/Controllers/Api/TransactionController.php`
- Method: `history()`
- Rate Limit: 200 requests per minute per IP

**Go:**
- File: `handlers/transaction_handler.go`
- Function: `History()`
- Rate Limit: 200 requests per minute per IP

**Request Body:** None (Bearer Token required)

**Success Response (200):**
```json
{
    "status": "success",
    "data": [
        {
            "id": 42,
            "transaction_date": "18-12-2026 13:22",
            "supplier_code": "7600",
            "lp_name": "ARYOS",
            "route_no": "RWC-09D",
            "status_truck": "IN FROM WJ",
            "no_truck_epc": "AF25110000000003",
            "ritase": 1,
            "scanned_items": [
                {"epc_code": "BB25110000000006", "type": "Pallet"},
                {"epc_code": "BB25110000000131", "type": "Box"}
            ]
        }
    ]
}
```

**Differences:** None

---

### 5. GET /api/master-data

**Laravel:**
- File: `app/Http/Controllers/Api/TransactionController.php`
- Method: `masterData()`
- Rate Limit: 200 requests per minute per IP

**Go:**
- File: `handlers/master_data_handler.go`
- Function: `MasterData()`
- Rate Limit: 200 requests per minute per IP

**Request Body:** None (Bearer Token required)

**Success Response (200):**
```json
{
    "status": "success",
    "data": {
        "suppliers": [
            {"code": "7600", "name": "PT Jaya Makmur", "lp_name": "ARYOS"}
        ],
        "routes": [
            {"code": "RWC-09D", "description": "Karawang - Pulogadung"}
        ]
    }
}
```

**Differences:** None

---

### 6. GET /api/master-rfid

**Laravel:**
- File: `app/Http/Controllers/Api/MasterDataController.php`
- Method: `index()`
- Rate Limit: 200 requests per minute per IP

**Go:**
- File: `handlers/master_data_handler.go`
- Function: `MasterRfid()`
- Rate Limit: 200 requests per minute per IP

**Request Body:** None (Bearer Token required)

**Success Response (200):**
```json
{
    "status": "success",
    "data": [
        {"epc_code": "AF25110000000003", "category": "Truck", "description": "Truck B 1009"},
        {"epc_code": "BB25110000000006", "category": "Pallet", "description": "Pallet 1"},
        {"epc_code": "BB25110000000131", "category": "Box", "description": "Box 1"}
    ]
}
```

**Differences:** None

---

### 7. GET /api/health (Go only)

**Go:**
- File: `handlers/health_handler.go`
- Function: `Health()`

**Request Body:** None

**Success Response (200):**
```json
{
    "success": true,
    "message": "Go API is running"
}
```

---

## Database Tables Used

| Table | Used By |
|-------|---------|
| `users` | Login, Transaction Submit, Transaction History |
| `personal_access_tokens` | Login (create), Logout (delete), Protected endpoints (validate) |
| `master_suppliers` | Master Data, Transaction Submit |
| `master_routes` | Master Data, Transaction Submit |
| `master_rfid_tags` | Master RFID, Transaction Submit |
| `milk_run_logs` | Transaction Submit, Transaction History |
| `scanned_items` | Transaction Submit, Transaction History |
| `model_has_roles` | Login (get role name) |

---

## Rate Limiting Comparison

| Route Group | Laravel | Go |
|-------------|---------|-----|
| Login | `throttle:7,1` (7 req/min/IP) | 7 req/min per IP |
| Protected | `throttle:200,1` (200 req/min/IP) | 200 req/min per IP |

---

## Status Summary

| Endpoint | Status | Notes |
|----------|--------|-------|
| `/api/health` | ✅ Done | Go-specific internal endpoint |
| `/api/login` | ✅ Done | Sanctum-compatible token generation |
| `/api/logout` | ✅ Done | Token revocation |
| `/api/transactions/submit` | ✅ Done | Full validation + DB transaction |
| `/api/transactions/history` | ✅ Done | 15 latest transactions with scanned items |
| `/api/master-data` | ✅ Done | Suppliers and routes |
| `/api/master-rfid` | ✅ Done | All RFID tags |

All endpoints are implemented with 1:1 compatibility with Laravel API.