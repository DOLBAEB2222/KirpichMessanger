# Telegram Clone - Optimized Messenger

A high-performance, resource-efficient messenger application built with Go, optimized for low-end hardware while maintaining scalability for 500+ users.

## 🚀 Features

### Core Features (MVP)
- ✅ **User Authentication** - JWT-based secure authentication
- ✅ **Direct Messages (DM)** - One-on-one conversations with automatic chat creation
- ✅ **Group Chats** - Multi-user group messaging
- ✅ **Channels** - Broadcast channels for announcements
- ✅ **Real-time Messaging** - WebSocket-based instant messaging with Redis pub/sub
- ✅ **Media Sharing** - Image, video, audio, and file uploads with compression
- ✅ **Premium Subscriptions** - Tiered subscription system

### DM Features (Stage 3)
- ✅ **Get or Create DM** - `GET /chats/dm/:user_id` endpoint for quick DM access
- ✅ **DM Uniqueness** - Only one DM chat exists between any two users (enforced via SQL constraints)
- ✅ **DM Caching** - Redis cache for DM lookups with 5-minute TTL
- ✅ **Auto-named Chats** - DM chats automatically named after the other user
- ✅ **Read Receipts** - Real-time read status updates via WebSocket
- ✅ **Typing Indicators** - Real-time typing status with 3-second debounce
- ✅ **Online Status** - User presence tracking via WebSocket

### WebSocket Features (Stage 3-4)
- ✅ **Real-time Events** - Bidirectional messaging via WebSocket
- ✅ **Typing Events** - Broadcast typing indicators to chat members
- ✅ **Read Receipts** - Notify when messages are read
- ✅ **Online Status** - Track and broadcast user presence
- ✅ **Chat Presence** - Join/leave notifications
- ✅ **Redis Pub/Sub** - Scalable message broadcasting
- ✅ **Automatic Reconnection** - Ping/pong keep-alive mechanism
- ✅ **WebRTC Signaling** - offer/answer/ice candidate exchange for calls

### Media Features (Stage 3)
- ✅ **Image Compression** - Automatic resizing to max 500px width
- ✅ **Adaptive Quality** - Quality adjusts based on original file size (70-85%)
- ✅ **Thumbnail Generation** - 200px thumbnails for image previews
- ✅ **File Validation** - MIME type, extension, and size validation
- ✅ **Path Traversal Protection** - Secure filename handling
- ✅ **Organized Storage** - Date-based directory structure (`uploads/2026/01/15/`)
- ✅ **Upload Rate Limiting** - 10 uploads per hour per user
- ✅ **Size Limits** - 50MB max per file (MVP)
- ✅ **Supported Types**: JPEG, PNG, GIF, WebP, MP4, WebM, MP3, WAV, PDF, ZIP, TXT
- ✅ **Media Cleanup** - Automatic removal of files older than 30 days

### Chat List Optimization (Stage 3)
- ✅ **Last Message Loading** - Chats include most recent message
- ✅ **Unread Counts** - Real-time unread message counts per chat
- ✅ **Redis Caching** - 5-minute cache for chat lists
- ✅ **Cache Invalidation** - Automatic invalidation on new messages
- ✅ **Efficient Queries** - Optimized SQL with proper indexing

### Premium Features
- Higher upload limits (500MB vs 50MB)
- Increased rate limits (1000 req/min vs 100 req/min)
- Priority support
- Custom themes (future)
- Advanced features (future)

### Voice & Video Calls (Stage 4)
- ✅ **Call Initiation** - `POST /api/v1/calls` to start voice/video calls
- ✅ **Call Signaling** - WebRTC signaling via WebSocket
- ✅ **ICE Servers** - TURN/STUN server configuration endpoint
- ✅ **Call Management** - Accept, reject, and end calls via REST API
- ✅ **Call History** - Persistent call records with duration

### Upcoming Features
- 🔜 **E2E Encryption** - End-to-end encryption (Stage 3-4)
- 🔜 **Real Payment Integration** - Stripe/Yookassa (Stage 2)

## 📋 MVP Notes

### Payment System (Current Implementation)
⚠️ **Important:** The MVP uses a **stub payment system**:
- No real payment processing
- No credit card validation
- Payments are logged to database only
- `is_premium` status is activated immediately upon "purchase"
- Real Stripe/Yookassa integration planned for **Stage 2**

**Example Response:**
```json
{
  "success": true,
  "message": "MVP: Payment stub activated. No real charge applied.",
  "subscription": { ... }
}
```

### Security (Current Implementation)
- ✅ HTTPS transport encryption via Caddy
- ✅ bcrypt password hashing (cost=12)
- ✅ JWT token-based authentication
- ❌ E2E encryption (planned for Stage 3-4)

Messages are currently stored in plaintext in the database. End-to-end encryption using Signal Protocol or similar will be implemented in **Stage 3-4**.

## 🏗️ Architecture

### Technology Stack
- **Backend:** Go 1.21+ with Fiber v3
- **Database:** PostgreSQL 16
- **Cache:** Redis 7
- **Reverse Proxy:** Caddy 2
- **WebRTC:** coturn TURN server
- **Containerization:** Docker & Docker Compose

### System Requirements

#### Minimum (MVP Deployment)
- **CPU:** Intel i3-2120 (2 cores, 3.3GHz) or equivalent
- **RAM:** 4GB DDR3
- **Storage:** 20GB SSD
- **Network:** 10 Mbps upload
- **OS:** Ubuntu 24.04 LTS

#### Recommended (Production)
- **CPU:** 4+ cores
- **RAM:** 8GB+
- **Storage:** 100GB+ SSD
- **Network:** 100 Mbps+

### Resource Allocation
| Service    | Memory Limit | Purpose                       |
|------------|--------------|-------------------------------|
| PostgreSQL | 800MB        | Primary database              |
| Redis      | 300MB        | Cache & pub/sub               |
| Go API     | 1GB          | REST API & WebSocket          |
| Caddy      | 200MB        | Reverse proxy & HTTPS         |
| coturn     | 200MB        | TURN server for WebRTC        |

## 🚀 Quick Start

### Automated Setup (Recommended)
```bash
# Download and run setup script
bash <(curl -fsSL https://raw.githubusercontent.com/your-repo/messenger/main/deploy/setup.sh)
```

The script will:
1. Check system requirements
2. Install Docker, Go, and dependencies
3. Configure firewall
4. Generate environment files with secure secrets
5. Start all services

### Manual Setup

#### 1. Clone Repository
```bash
git clone https://github.com/your-repo/messenger.git
cd messenger
```

#### 2. Configure Environment
```bash
cp deploy/.env.example deploy/.env
nano deploy/.env
```

**Important:** Update these values in `.env`:
- `JWT_SECRET` - Generate with: `openssl rand -hex 32`
- `POSTGRES_PASSWORD` - Strong database password
- `TURN_PASSWORD` - TURN server password
- `DOMAIN` - Your actual domain (for production)

#### 3. Start Services
```bash
cd deploy
docker compose up -d
```

#### 4. Verify Deployment
```bash
# Check service status
docker compose ps

# View logs
docker compose logs -f api

# Test API health
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "timestamp": 1705320000,
  "service": "messenger-api"
}
```

### Development Setup

#### Prerequisites
- Go 1.21+
- PostgreSQL 16
- Redis 7

#### Run Locally
```bash
cd backend
cp .env.example .env
go mod download
go run cmd/api/main.go
```

## 📖 API Documentation

### Authentication

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "phone": "+1234567890",
  "email": "user@example.com",
  "password": "SecurePass123!",
  "username": "johndoe"
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "phone": "+1234567890",
  "password": "SecurePass123!"
}
```

### Messaging

#### Send Message
```http
POST /api/v1/messages
Authorization: Bearer <token>
Content-Type: application/json

{
  "chat_id": "uuid",
  "content": "Hello, world!",
  "message_type": "text"
}
```

#### Send Media Message
```http
POST /api/v1/messages/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data

chat_id: "uuid"
file: <binary file>
content: "Optional caption"
```

#### Get Chat Messages
```http
GET /api/v1/chats/:chatId/messages?limit=50&offset=0
Authorization: Bearer <token>
```

#### Mark Chat as Read
```http
POST /api/v1/chats/:chatId/read
Authorization: Bearer <token>
```

### DM

#### Get or Create DM Chat
```http
GET /api/v1/chats/dm/:user_id
Authorization: Bearer <token>

# Returns existing chat or creates new one
# Response (200 or 201):
{
  "id": "uuid",
  "name": "other_user_username",
  "type": "dm",
  "member_count": 2
}
```

### Subscriptions (MVP - Stub Payment)

#### Purchase Subscription
```http
POST /api/v1/subscriptions/purchase
Authorization: Bearer <token>
Content-Type: application/json

{
  "subscription_type": "premium_monthly"
}
```

**Response:**
```json
{
  "success": true,
  "message": "MVP: Payment stub activated. No real charge applied.",
  "subscription": {
    "id": "uuid",
    "type": "premium_monthly",
    "end_date": "2024-02-15"
  }
}
```

### WebSocket Real-time Events

Connect to WebSocket for real-time updates:
```
WS /ws?token=<jwt_token>
```

**Client → Server:**
```json
// Send message
{ "type": "message", "chat_id": "uuid", "content": "Hello" }

// Typing indicator
{ "type": "typing", "chat_id": "uuid" }

// Mark as read
{ "type": "read", "chat_id": "uuid" }

// Join chat for presence
{ "type": "join_chat", "chat_id": "uuid" }

// WebRTC Offer (call signaling)
{ "type": "call:webrtc:offer", "data": { "call_id": "uuid", "offer": { ... } } }

// WebRTC Answer
{ "type": "call:webrtc:answer", "data": { "call_id": "uuid", "answer": { ... } } }

// ICE Candidate
{ "type": "call:webrtc:candidate", "data": { "call_id": "uuid", "candidate": { ... } } }

// Hang up
{ "type": "call:hangup", "data": { "call_id": "uuid" } }
```

**Server → Client:**
```json
// New message
{ "type": "new_message", "message": { ... } }

// Typing status
{ "type": "typing", "chat_id": "uuid", "user_id": "uuid", "is_typing": true }

// Read receipt
{ "type": "read", "chat_id": "uuid", "user_id": "uuid", "unread_count": 0 }

// Online status
{ "type": "online_status", "user_id": "uuid", "is_online": true }

// Call initiated
{ "type": "call:initiate", "call": { ... }, "initiator": { ... } }

// Call accepted/rejected
{ "type": "call:accepted", "call": { ... } }
{ "type": "call:rejected", "call": { ... } }

// Call ended
{ "type": "call:ended", "call": { ... }, "ended_by": "uuid" }
```

See [TDD.md](docs/TDD.md) for complete API documentation.

## 🛠️ Development

### Build Backend
```bash
cd backend
go build -o bin/api cmd/api/main.go
```

### Run Tests
```bash
go test ./...
```

### Database Migrations
```bash
# Migrations run automatically on startup
# Manual migration:
docker exec -it messenger-postgres psql -U messenger -d messenger -f /schema.sql
```

### Useful Commands
```bash
# View logs
docker compose logs -f api

# Restart services
docker compose restart

# Stop services
docker compose down

# Database access
docker exec -it messenger-postgres psql -U messenger -d messenger

# Redis CLI
docker exec -it messenger-redis redis-cli
```

## 🔒 Security

### Current Implementation (MVP)
- HTTPS transport encryption (Caddy with Let's Encrypt)
- bcrypt password hashing (cost=12)
- JWT authentication (HS256)
- Input validation and sanitization
- Rate limiting

### Future Implementation (Stage 3-4)
- End-to-end encryption (Signal Protocol)
- Perfect forward secrecy
- Client-side key generation
- Two-factor authentication (2FA)

## 🌐 Deployment

### Production Checklist
- [ ] Update `DOMAIN` in `.env` to actual domain
- [ ] Set strong `JWT_SECRET` (32+ characters)
- [ ] Configure SSL certificate (Caddy handles automatically)
- [ ] Set up database backups
- [ ] Configure firewall rules
- [ ] Set up monitoring (Prometheus/Grafana recommended)
- [ ] Enable log rotation
- [ ] Configure email notifications

### Firewall Ports
```bash
sudo ufw allow 22/tcp   # SSH
sudo ufw allow 80/tcp   # HTTP
sudo ufw allow 443/tcp  # HTTPS
sudo ufw allow 3478:3479/tcp  # TURN
sudo ufw allow 3478:3479/udp  # TURN
sudo ufw allow 49152:49252/udp  # TURN relay
```

## 📊 Monitoring

### Health Checks
- API: `GET /health`
- PostgreSQL: `pg_isready`
- Redis: `PING`

### Metrics
- CPU usage per container
- Memory usage per container
- Database connection count
- WebSocket connections
- API response times

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## 📅 Roadmap

See [ROADMAP.md](docs/ROADMAP.md) for detailed development stages.

**Quick Overview:**
- **Stage 1 (MVP):** Core messaging, stub payments ✅
- **Stage 2:** Real payment integration, media optimization 🔜
- **Stage 3:** E2E encryption, voice/video calls 🔜
- **Stage 4:** Advanced features, horizontal scaling 🔜

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For issues and questions:
- Open an issue on GitHub
- Check documentation in `/docs`
- Review [TDD.md](docs/TDD.md) for technical details

## ⚠️ Disclaimer

**This is MVP software:**
- Payment system is stubbed (no real charges)
- No end-to-end encryption (planned for later)
- Optimized for low-end hardware but ready to scale
- Not production-ready until Stage 2+

For production deployment with real payments and E2E encryption, wait for Stage 2-3 releases.

---

**Version:** 1.0.0 (MVP)  
**Last Updated:** 2024-01-15
