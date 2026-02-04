# API Documentation

## Base URL

```
Production: https://yourdomain.com/api/v1
Development: http://localhost:8080/api/v1
```

## Authentication

All protected endpoints require an `Authorization` header with a Bearer token:

```
Authorization: Bearer <jwt_token>
```

## Endpoints

### Authentication

#### POST /auth/register

Register a new user with phone number or email.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+79991234567",
    "email": "user@example.com",
    "password": "SecurePass123",
    "username": "john_doe"
  }'
```

**Response (201):**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}
```

**Validation Rules:**
- `phone`: Required if email not provided. E.164 format (e.g., +79991234567)
- `email`: Optional. Valid email format
- `password`: Required. Min 8 chars, must contain uppercase, lowercase, and digit
- `username`: Optional. 3-50 chars, alphanumeric and underscores only

---

#### POST /auth/login

Login with phone or email + password.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone_or_email": "+79991234567",
    "password": "SecurePass123"
  }'
```

**Response (200):**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}
```

**Rate Limiting:** Max 5 attempts per 15 minutes per phone/email.

---

#### POST /auth/refresh

Refresh access token using refresh token.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}
```

---

#### POST /auth/logout

Logout and invalidate the current token.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <jwt_token>"
```

**Response (200):**
```json
{
  "message": "Logged out successfully"
}
```

---

### User Profile

#### GET /users/me

Get current user's full profile.

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <jwt_token>"
```

**Response (200):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "phone": "+79991234567",
  "email": "user@example.com",
  "username": "john_doe",
  "avatar_url": "https://cdn.example.com/avatars/user.jpg",
  "bio": "Hello, I'm using Messenger!",
  "created_at": "2026-02-02T10:00:00Z",
  "is_premium": false,
  "last_seen": "2026-02-02T17:00:00Z"
}
```

---

#### GET /users/:user_id

Get public profile of another user.

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/users/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer <jwt_token>"
```

**Response (200):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "john_doe",
  "avatar_url": "https://cdn.example.com/avatars/user.jpg",
  "bio": "Hello, I'm using Messenger!",
  "is_premium": false,
  "last_seen": "2026-02-02T17:00:00Z"
}
```

**Note:** This endpoint returns limited information (no phone/email).

---

#### PATCH /users/me

Update current user's profile.

**Request:**
```bash
curl -X PATCH http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "new_username",
    "bio": "New bio here",
    "avatar": "https://cdn.example.com/avatars/new.jpg"
  }'
```

**Response (200):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "new_username",
  "bio": "New bio here",
  "avatar_url": "https://cdn.example.com/avatars/new.jpg",
  "updated_at": "2026-02-02T18:00:00Z"
}
```

**Validation Rules:**
- `username`: 3-50 chars, alphanumeric and underscores only
- `bio`: Max 500 characters
- `avatar`: URL or base64 encoded image

---

#### PATCH /users/me/password

Change current user's password.

**Request:**
```bash
curl -X PATCH http://localhost:8080/api/v1/users/me/password \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "currentpassword",
    "new_password": "newpassword"
  }'
```

**Response (200):**
```json
{
  "message": "Password changed successfully"
}
```

**Validation Rules:**
- `new_password`: Min 8 chars, must contain uppercase, lowercase, and digit

---

#### DELETE /users/me

Delete current user's account (soft delete).

**Request:**
```bash
curl -X DELETE http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <jwt_token>"
```

**Response (200):**
```json
{
  "message": "Account deleted successfully"
}
```

---

## Error Responses

All errors follow this format:

```json
{
  "error": "Error description"
}
```

### HTTP Status Codes

| Status | Description |
|--------|-------------|
| 200 OK | Request successful |
| 201 Created | Resource created successfully |
| 400 Bad Request | Invalid request data |
| 401 Unauthorized | Missing or invalid token |
| 403 Forbidden | Not allowed to access resource |
| 404 Not Found | Resource not found |
| 409 Conflict | Resource already exists |
| 429 Too Many Requests | Rate limit exceeded |
| 500 Internal Server Error | Server error |

### Error Examples

**400 Bad Request:**
```json
{
  "errors": {
    "password": "Password must be at least 8 characters long",
    "email": "Invalid email format"
  }
}
```

**401 Unauthorized:**
```json
{
  "error": "Invalid or expired token"
}
```

**409 Conflict:**
```json
{
  "error": "Phone number already registered"
}
```

**429 Too Many Requests:**
```json
{
  "error": "Too many login attempts",
  "retry_after": 900
}
```

---

## Rate Limiting

| Endpoint | Limit |
|----------|-------|
| POST /auth/login | 5 requests per 15 minutes per phone/email |
| POST /auth/register | 10 requests per hour per IP |
| All other endpoints | 100 requests per minute per user |

---

## Security

### Password Requirements
- Minimum 8 characters
- At least one uppercase letter (A-Z)
- At least one lowercase letter (a-z)
- At least one digit (0-9)

### Token Specifications
- **Access Token:** JWT, 1 hour expiration
- **Refresh Token:** JWT, 30 days expiration
- **Algorithm:** HS256
- **Secret:** Environment variable `JWT_SECRET`

### Password Hashing
- **Algorithm:** bcrypt
- **Cost Factor:** 12 rounds

---

## Testing with curl

### Complete Authentication Flow

1. **Register:**
```bash
# Register with phone
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+79991234567",
    "password": "TestPass123",
    "username": "testuser"
  }'

# Register with email
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123",
    "username": "testuser"
  }'
```

2. **Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone_or_email": "+79991234567",
    "password": "TestPass123"
  }'
```

3. **Get Profile:**
```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

4. **Update Profile:**
```bash
curl -X PATCH http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "bio": "My new bio"
  }'
```

5. **Refresh Token:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN_HERE"
  }'
```

6. **Logout:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

## WebSocket Connection

For real-time messaging, connect to WebSocket endpoint:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws?token=YOUR_JWT_TOKEN');

ws.onopen = () => {
  console.log('Connected to WebSocket');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected from WebSocket');
};
```

---

## Postman Collection

You can import the following collection for easy testing:

```json
{
  "info": {
    "name": "Messenger API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Auth",
      "item": [
        {
          "name": "Register",
          "request": {
            "method": "POST",
            "header": [{"key": "Content-Type", "value": "application/json"}],
            "url": "{{base_url}}/auth/register",
            "body": {
              "mode": "raw",
              "raw": "{\"phone\": \"+79991234567\", \"password\": \"TestPass123\", \"username\": \"testuser\"}"
            }
          }
        },
        {
          "name": "Login",
          "request": {
            "method": "POST",
            "header": [{"key": "Content-Type", "value": "application/json"}],
            "url": "{{base_url}}/auth/login",
            "body": {
              "mode": "raw",
              "raw": "{\"phone_or_email\": \"+79991234567\", \"password\": \"TestPass123\"}"
            }
          }
        },
        {
          "name": "Refresh Token",
          "request": {
            "method": "POST",
            "header": [{"key": "Content-Type", "value": "application/json"}],
            "url": "{{base_url}}/auth/refresh",
            "body": {
              "mode": "raw",
              "raw": "{\"refresh_token\": \"{{refresh_token}}\"}"
            }
          }
        },
        {
          "name": "Logout",
          "request": {
            "method": "POST",
            "header": [{"key": "Authorization", "value": "Bearer {{token}}"}],
            "url": "{{base_url}}/auth/logout"
          }
        }
      ]
    },
    {
      "name": "Users",
      "item": [
        {
          "name": "Get Me",
          "request": {
            "method": "GET",
            "header": [{"key": "Authorization", "value": "Bearer {{token}}"}],
            "url": "{{base_url}}/users/me"
          }
        },
        {
          "name": "Get User",
          "request": {
            "method": "GET",
            "header": [{"key": "Authorization", "value": "Bearer {{token}}"}],
            "url": "{{base_url}}/users/{{user_id}}"
          }
        },
        {
          "name": "Update Profile",
          "request": {
            "method": "PATCH",
            "header": [
              {"key": "Authorization", "value": "Bearer {{token}}"},
              {"key": "Content-Type", "value": "application/json"}
            ],
            "url": "{{base_url}}/users/me",
            "body": {
              "mode": "raw",
              "raw": "{\"bio\": \"My updated bio\"}"
            }
          }
        },
        {
          "name": "Change Password",
          "request": {
            "method": "PATCH",
            "header": [
              {"key": "Authorization", "value": "Bearer {{token}}"},
              {"key": "Content-Type", "value": "application/json"}
            ],
            "url": "{{base_url}}/users/me/password",
            "body": {
              "mode": "raw",
              "raw": "{\"old_password\": \"TestPass123\", \"new_password\": \"NewPass123\"}"
            }
          }
        },
        {
          "name": "Delete Account",
          "request": {
            "method": "DELETE",
            "header": [{"key": "Authorization", "value": "Bearer {{token}}"}],
            "url": "{{base_url}}/users/me"
          }
        }
      ]
    }
  ],
  "variable": [
    {"key": "base_url", "value": "http://localhost:8080/api/v1"},
    {"key": "token", "value": ""},
    {"key": "refresh_token", "value": ""},
    {"key": "user_id", "value": ""}
  ]
}
```

---

**Last Updated:** 2026-02-02  
**API Version:** v1.0.0
