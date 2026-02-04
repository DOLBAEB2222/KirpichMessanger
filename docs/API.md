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

### Messages

#### POST /messages

Send a text message to a chat.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "chat_id": "550e8400-e29b-41d4-a716-446655440000",
    "content": "Hello, world!",
    "message_type": "text",
    "reply_to_id": "optional-message-id"
  }'
```

**Response (201):**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "sender_id": "550e8400-e29b-41d4-a716-446655440000",
  "chat_id": "550e8400-e29b-41d4-a716-446655440001",
  "content": "Hello, world!",
  "message_type": "text",
  "is_edited": false,
  "created_at": "2026-02-02T17:00:00Z",
  "sender": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "john_doe"
  }
}
```

---

#### POST /messages/upload

Upload media (image, video, audio, file) to a chat.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/messages/upload \
  -H "Authorization: Bearer <jwt_token>" \
  -F "chat_id=550e8400-e29b-41d4-a716-446655440000" \
  -F "content=Optional caption" \
  -F "file=@/path/to/image.jpg"
```

**Response (201):**
```json
{
  "message": {
    "id": "660e8400-e29b-41d4-a716-446655440002",
    "sender_id": "550e8400-e29b-41d4-a716-446655440000",
    "chat_id": "550e8400-e29b-41d4-a716-446655440001",
    "content": "image.jpg",
    "message_type": "image",
    "media_url": "2026/01/15/image_abc123.jpg",
    "created_at": "2026-02-02T17:00:00Z"
  },
  "media": {
    "file_path": "2026/01/15/image_abc123.jpg",
    "file_size": 204800,
    "mime_type": "image/jpeg",
    "width": 800,
    "height": 600,
    "thumbnail": "2026/01/15/thumb_image_abc123.jpg",
    "compressed": true
  }
}
```

**Supported File Types:**
- Images: jpeg, jpg, png, gif, webp (max 50MB)
- Videos: mp4, webm, ogg (max 50MB)
- Audio: mp3, wav, ogg, webm (max 50MB)
- Files: pdf, zip, txt (max 50MB)

**Rate Limiting:** Max 10 uploads per hour per user.

---

#### GET /messages/:id

Get a specific message by ID.

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/messages/660e8400-e29b-41d4-a716-446655440001 \
  -H "Authorization: Bearer <jwt_token>"
```

**Response (200):**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "sender_id": "550e8400-e29b-41d4-a716-446655440000",
  "chat_id": "550e8400-e29b-41d4-a716-446655440001",
  "content": "Hello, world!",
  "message_type": "text",
  "is_edited": false,
  "created_at": "2026-02-02T17:00:00Z",
  "sender": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "john_doe"
  }
}
```

---

#### PATCH /messages/:id

Edit a message (only sender can edit).

**Request:**
```bash
curl -X PATCH http://localhost:8080/api/v1/messages/660e8400-e29b-41d4-a716-446655440001 \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Updated message content"
  }'
```

**Response (200):** Updated message object

---

#### DELETE /messages/:id

Delete a message (soft delete, only sender can delete).

**Request:**
```bash
curl -X DELETE http://localhost:8080/api/v1/messages/660e8400-e29b-41d4-a716-446655440001 \
  -H "Authorization: Bearer <jwt_token>"
```

**Response (200):**
```json
{
  "message": "Message deleted successfully"
}
```

---

### Chats

#### GET /chats

Get all chats for the current user with last messages and unread counts.

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/chats \
  -H "Authorization: Bearer <jwt_token>"
```

**Response (200):**
```json
{
  "chats": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "John Doe",
      "type": "dm",
      "member_count": 2,
      "last_message_at": "2026-02-02T17:00:00Z",
      "last_message": {
        "id": "660e8400-e29b-41d4-a716-446655440001",
        "content": "Hello!",
        "message_type": "text",
        "sender": {
          "id": "550e8400-e29b-41d4-a716-446655440000",
          "username": "john_doe"
        }
      },
      "unread_count": 3
    }
  ],
  "count": 1
}
```

**Caching:** Results are cached for 5 minutes and invalidated on new messages.

---

#### POST /chats

Create a new chat (group) or get existing DM.

**Request (Group):**
```bash
curl -X POST http://localhost:8080/api/v1/chats \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Project Team",
    "type": "group",
    "member_ids": ["user-id-1", "user-id-2"]
  }'
```

**Request (DM):**
```bash
curl -X POST http://localhost:8080/api/v1/chats \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "dm",
    "member_ids": ["user-id-1"]
  }'
```

**Response (201):** Chat object

---

#### GET /chats/dm/:user_id

Get or create a DM chat with a specific user.

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/chats/dm/550e8400-e29b-41d4-a716-446655440002 \
  -H "Authorization: Bearer <jwt_token>"
```

**Response (200):** Chat object (existing or newly created)

---

#### GET /chats/:id

Get a specific chat by ID.

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/chats/550e8400-e29b-41d4-a716-446655440001 \
  -H "Authorization: Bearer <jwt_token>"
```

**Response (200):** Chat object with members

---

#### GET /chats/:id/messages

Get messages from a chat with pagination.

**Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/chats/550e8400-e29b-41d4-a716-446655440001/messages?limit=50&offset=0" \
  -H "Authorization: Bearer <jwt_token>"
```

**Response (200):**
```json
{
  "messages": [...],
  "total": 100,
  "limit": 50,
  "offset": 0,
  "has_more": true
}
```

---

#### POST /chats/:id/read

Mark all messages in a chat as read.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/chats/550e8400-e29b-41d4-a716-446655440001/read \
  -H "Authorization: Bearer <jwt_token>"
```

**Response (200):**
```json
{
  "message": "Marked as read",
  "unread_count": 0
}
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

### WebSocket Event Types

#### Client → Server

**Send Message:**
```json
{
  "type": "message",
  "chat_id": "550e8400-e29b-41d4-a716-446655440001",
  "content": "Hello via WebSocket!"
}
```

**Typing Indicator:**
```json
{
  "type": "typing",
  "chat_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

**Read Receipt:**
```json
{
  "type": "read",
  "chat_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

**Join Chat:**
```json
{
  "type": "join_chat",
  "chat_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

**Leave Chat:**
```json
{
  "type": "leave_chat",
  "chat_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

**Ping (keep-alive):**
```json
{
  "type": "ping"
}
```

#### Server → Client

**New Message:**
```json
{
  "type": "new_message",
  "message": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "sender_id": "550e8400-e29b-41d4-a716-446655440000",
    "chat_id": "550e8400-e29b-41d4-a716-446655440001",
    "content": "Hello!",
    "message_type": "text",
    "created_at": "2026-02-02T17:00:00Z"
  }
}
```

**Typing Indicator:**
```json
{
  "type": "typing",
  "chat_id": "550e8400-e29b-41d4-a716-446655440001",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "is_typing": true,
  "timestamp": 1738515600
}
```

**Read Receipt:**
```json
{
  "type": "read",
  "chat_id": "550e8400-e29b-41d4-a716-446655440001",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "last_read_at": "2026-02-02T17:00:00Z",
  "unread_count": 0,
  "message_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

**Online Status:**
```json
{
  "type": "online_status",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "is_online": true,
  "timestamp": 1738515600
}
```

**Chat Presence:**
```json
{
  "type": "chat_presence",
  "chat_id": "550e8400-e29b-41d4-a716-446655440001",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "is_joined": true
}
```

**Pong (ping response):**
```json
{
  "type": "pong",
  "timestamp": 1738515600
}
```

**User Chats (sent on connect):**
```json
{
  "type": "user_chats",
  "chat_ids": ["chat-id-1", "chat-id-2"]
}
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
