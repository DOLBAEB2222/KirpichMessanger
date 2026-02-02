# Technical Design Document (TDD)

## Project Overview

**Name:** Telegram Clone - Optimized Messenger  
**Version:** 1.0.0 (MVP)  
**Target Platform:** Ubuntu 24.04 LTS  
**Target Hardware:** Intel i3-2120, 4GB RAM (optimized for low-end, scalable to 500+ users)

## Executive Summary

This messenger application is designed as a scalable Telegram clone with a focus on resource efficiency. The MVP includes core messaging features, group chats, channels, and a **stub payment system** (no real payment processing). End-to-end encryption (E2E) is planned for stage 3-4, not included in MVP.

### Key MVP Limitations & Future Plans
1. **Payment System (MVP):** Stub implementation - logs payment to database, activates premium without card validation. Real Stripe/Yookassa integration planned for stage 2.
2. **E2E Encryption:** Planned for stage 3-4. MVP uses HTTPS transport security only.
3. **Scalability:** Architecture supports 500+ users and horizontal scaling, but optimized for 4GB RAM deployment.

---

## System Architecture

### Component Diagram (Text)

```
┌─────────────────────────────────────────────────────────────┐
│                         Internet                             │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   Caddy (Reverse Proxy)                      │
│            HTTPS Termination, Static Files                   │
│                    Ports: 80, 443                            │
└───────────┬──────────────────────────────┬──────────────────┘
            │                              │
            ▼                              ▼
┌───────────────────────┐      ┌──────────────────────────────┐
│   Go Fiber API        │      │   WebSocket Handler          │
│   REST + WebSocket    │      │   Real-time Messages         │
│   Port: 8080          │      │   (Same Process)             │
│   Memory: 1GB         │      └──────────────────────────────┘
└───────┬───────────────┘
        │
        ├──────────┬──────────────┬──────────────────┐
        ▼          ▼              ▼                  ▼
┌──────────┐ ┌──────────┐ ┌──────────────┐ ┌──────────────────┐
│PostgreSQL│ │  Redis   │ │  File System │ │   coturn (TURN)  │
│  16      │ │  7       │ │  (Media)     │ │   WebRTC Server  │
│Port: 5432│ │Port: 6379│ │  /data/media │ │   Ports: 3478-79 │
│Mem: 800MB│ │Mem: 300MB│ └──────────────┘ │   Mem: 200MB     │
└──────────┘ └──────────┘                  └──────────────────┘
```

### Data Flow

1. **Authentication:** Client → Caddy → API → PostgreSQL (JWT issued)
2. **Messaging:** Client → WebSocket → Redis Pub/Sub → PostgreSQL → Recipients
3. **Media Upload:** Client → API → File System → PostgreSQL (metadata)
4. **Premium Subscription (MVP):** Client → API → payment_logs table (stub) → User.is_premium = true
5. **Voice/Video Calls:** Client ↔ coturn (TURN) ↔ Client (P2P via WebRTC)

---

## Resource Requirements

### Target Server (MVP Deployment)
- **CPU:** Intel i3-2120 (2 cores, 4 threads @ 3.3GHz) or equivalent
- **RAM:** 4GB DDR3
- **Storage:** 20GB SSD (minimum), 100GB+ recommended for media
- **Network:** 10 Mbps upload (for 10-20 concurrent users)
- **OS:** Ubuntu 24.04 LTS

### Resource Allocation
| Service       | Memory Limit | CPU Priority | Notes                          |
|---------------|--------------|--------------|--------------------------------|
| PostgreSQL    | 800MB        | High         | shared_buffers=512MB           |
| Redis         | 300MB        | Medium       | LRU eviction policy            |
| Go API        | 1GB          | High         | Includes WebSocket connections |
| Caddy         | 200MB        | Low          | Lightweight proxy              |
| coturn        | 200MB        | Medium       | TURN server for WebRTC         |
| System        | 1.5GB        | -            | OS + overhead                  |
| **Total**     | **3.5GB**    | -            | 500MB reserved buffer          |

### Scalability Targets
- **Phase 1 (MVP):** 50-100 concurrent users, 500+ total users
- **Phase 2:** 200-500 concurrent users (add Redis Cluster, read replicas)
- **Phase 3:** 1000+ users (horizontal API scaling, CDN for media)

---

## Technology Stack

### Backend
- **Language:** Go 1.21+
  - *Why:* High performance, low memory footprint, excellent concurrency model
- **Framework:** Fiber v3
  - *Why:* Fastest Go web framework, low memory usage, Express-like API
- **ORM:** GORM v2
  - *Why:* Auto-migrations, relationship management, query optimization

### Database
- **Primary DB:** PostgreSQL 16
  - *Why:* JSONB support, full-text search, ACID compliance, mature ecosystem
- **Caching:** Redis 7
  - *Why:* In-memory speed, pub/sub for WebSocket, session storage

### Infrastructure
- **Containerization:** Docker + Docker Compose
  - *Why:* Reproducible deployments, resource isolation
- **Reverse Proxy:** Caddy 2
  - *Why:* Automatic HTTPS, minimal configuration, low memory usage
- **WebRTC:** coturn TURN server
  - *Why:* NAT traversal for voice/video calls

### Security
- **Authentication:** JWT (HS256)
- **Transport:** HTTPS via Caddy (Let's Encrypt)
- **Password Hashing:** bcrypt (cost=12)
- **E2E Encryption:** Planned for stage 3-4 (Signal Protocol or similar)

---

## API Contracts

### Base URL
```
Production: https://yourdomain.com/api/v1
Development: http://localhost:8080/api/v1
```

### Authentication Endpoints

#### 1. Register User
```http
POST /auth/register
Content-Type: application/json

Request:
{
  "phone": "+1234567890",
  "email": "user@example.com",
  "password": "SecurePass123!",
  "username": "johndoe"
}

Response (201):
{
  "user": {
    "id": "uuid",
    "phone": "+1234567890",
    "username": "johndoe",
    "created_at": "2024-01-15T10:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

#### 2. Login
```http
POST /auth/login
Content-Type: application/json

Request:
{
  "phone": "+1234567890",
  "password": "SecurePass123!"
}

Response (200):
{
  "user": {
    "id": "uuid",
    "username": "johndoe",
    "is_premium": false
  },
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### Messaging Endpoints

#### 3. Send Message
```http
POST /messages
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "chat_id": "uuid",
  "content": "Hello, world!",
  "message_type": "text"
}

Response (201):
{
  "id": "uuid",
  "sender_id": "uuid",
  "chat_id": "uuid",
  "content": "Hello, world!",
  "message_type": "text",
  "created_at": "2024-01-15T10:05:00Z"
}
```

#### 4. Get Chat Messages
```http
GET /chats/:chatId/messages?limit=50&offset=0
Authorization: Bearer <token>

Response (200):
{
  "messages": [
    {
      "id": "uuid",
      "sender_id": "uuid",
      "content": "Hello!",
      "created_at": "2024-01-15T10:00:00Z"
    }
  ],
  "total": 100,
  "has_more": true
}
```

### Chat/Group Endpoints

#### 5. Create Group
```http
POST /chats
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "name": "My Group",
  "type": "group",
  "member_ids": ["uuid1", "uuid2"]
}

Response (201):
{
  "id": "uuid",
  "name": "My Group",
  "type": "group",
  "owner_id": "uuid",
  "member_count": 3,
  "created_at": "2024-01-15T10:00:00Z"
}
```

#### 6. Create Channel
```http
POST /channels
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "name": "Announcements",
  "description": "Official announcements"
}

Response (201):
{
  "id": "uuid",
  "name": "Announcements",
  "owner_id": "uuid",
  "created_at": "2024-01-15T10:00:00Z"
}
```

### Subscription Endpoints (MVP - Stub Payment)

#### 7. Purchase Subscription
```http
POST /subscriptions/purchase
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "subscription_type": "premium_monthly",
  "payment_method": "stub"
}

Response (200):
{
  "subscription": {
    "id": "uuid",
    "user_id": "uuid",
    "type": "premium_monthly",
    "start_date": "2024-01-15",
    "end_date": "2024-02-15",
    "status": "active"
  },
  "payment_log": {
    "id": "uuid",
    "amount": 4.99,
    "status": "completed_stub"
  },
  "message": "MVP: Payment stub activated. No real charge."
}
```

#### 8. Get Subscription Status
```http
GET /subscriptions/me
Authorization: Bearer <token>

Response (200):
{
  "is_premium": true,
  "subscription": {
    "type": "premium_monthly",
    "end_date": "2024-02-15",
    "auto_renew": false
  }
}
```

### WebSocket Endpoint

#### 9. Real-time Connection
```
WS /ws
Authorization: Bearer <token> (via query param or header)

Client → Server Messages:
{
  "type": "message",
  "chat_id": "uuid",
  "content": "Hello"
}

Server → Client Messages:
{
  "type": "new_message",
  "message": {
    "id": "uuid",
    "sender_id": "uuid",
    "content": "Hello",
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

---

## Database Schema

### Entity Relationship Diagram (Conceptual)

```
users (1) ──────< (N) messages
users (1) ──────< (N) chat_members ────< (1) chats
users (1) ──────< (N) channels
users (1) ──────< (N) subscriptions
users (1) ──────< (N) payment_logs
```

### Table Definitions

#### users
```sql
- id: UUID PRIMARY KEY
- phone: VARCHAR(20) UNIQUE NOT NULL (indexed)
- email: VARCHAR(255) UNIQUE
- password_hash: VARCHAR(255) NOT NULL
- username: VARCHAR(50) UNIQUE
- avatar_url: TEXT
- bio: TEXT
- is_premium: BOOLEAN DEFAULT FALSE
- created_at: TIMESTAMP DEFAULT NOW()
- updated_at: TIMESTAMP DEFAULT NOW()
```

#### chats
```sql
- id: UUID PRIMARY KEY
- name: VARCHAR(255)
- type: ENUM('dm', 'group') NOT NULL
- owner_id: UUID REFERENCES users(id)
- created_at: TIMESTAMP DEFAULT NOW()
- updated_at: TIMESTAMP DEFAULT NOW()
```

#### chat_members
```sql
- chat_id: UUID REFERENCES chats(id) ON DELETE CASCADE
- user_id: UUID REFERENCES users(id) ON DELETE CASCADE
- role: ENUM('admin', 'member') DEFAULT 'member'
- joined_at: TIMESTAMP DEFAULT NOW()
- PRIMARY KEY (chat_id, user_id)
```

#### messages
```sql
- id: UUID PRIMARY KEY
- sender_id: UUID REFERENCES users(id) ON DELETE SET NULL
- chat_id: UUID REFERENCES chats(id) ON DELETE CASCADE
- content: TEXT NOT NULL
- message_type: ENUM('text', 'image', 'video', 'audio', 'file') DEFAULT 'text'
- media_url: TEXT
- created_at: TIMESTAMP DEFAULT NOW() (indexed DESC)
- updated_at: TIMESTAMP DEFAULT NOW()
```

#### channels
```sql
- id: UUID PRIMARY KEY
- name: VARCHAR(255) NOT NULL
- owner_id: UUID REFERENCES users(id) ON DELETE CASCADE
- description: TEXT
- subscriber_count: INTEGER DEFAULT 0
- created_at: TIMESTAMP DEFAULT NOW()
```

#### subscriptions
```sql
- id: UUID PRIMARY KEY
- user_id: UUID REFERENCES users(id) ON DELETE CASCADE
- type: ENUM('premium_monthly', 'premium_yearly') NOT NULL
- start_date: DATE NOT NULL
- end_date: DATE NOT NULL
- auto_renew: BOOLEAN DEFAULT FALSE
- status: ENUM('active', 'expired', 'cancelled') DEFAULT 'active'
- created_at: TIMESTAMP DEFAULT NOW()
```

#### payment_logs
```sql
- id: UUID PRIMARY KEY
- user_id: UUID REFERENCES users(id) ON DELETE SET NULL
- amount: DECIMAL(10, 2) NOT NULL
- subscription_type: VARCHAR(50) NOT NULL
- status: ENUM('completed_stub', 'pending', 'failed') DEFAULT 'pending'
- payment_method: VARCHAR(50) DEFAULT 'stub'
- notes: TEXT (for MVP: "Stub payment - no real charge")
- created_at: TIMESTAMP DEFAULT NOW()
```

### Indexes for Performance
```sql
CREATE INDEX idx_messages_chat_created ON messages(chat_id, created_at DESC);
CREATE INDEX idx_messages_sender ON messages(sender_id);
CREATE INDEX idx_chat_members_user ON chat_members(user_id);
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_subscriptions_user_status ON subscriptions(user_id, status);
```

---

## Optimization Strategies

### 1. Database Optimizations
- **Connection Pooling:** Max 100 connections, min 10 idle
- **Query Optimization:** Use EXPLAIN ANALYZE, add indexes for frequent queries
- **Partitioning:** Future - partition messages table by date for large datasets
- **Vacuuming:** Auto-vacuum enabled for PostgreSQL

### 2. Caching Strategy (Redis)
- **Session Storage:** JWT refresh tokens (TTL: 7 days)
- **User Profile Cache:** Hot user data (TTL: 1 hour)
- **Message Cache:** Recent messages per chat (TTL: 30 minutes)
- **Online Status:** User presence tracking (TTL: 5 minutes)
- **Pub/Sub:** Real-time message delivery via WebSocket

### 3. Media Optimization
- **Image Compression:** JPEG quality 85%, WebP format preferred
- **Video Compression:** H.264, max 720p for free users, 1080p for premium
- **Thumbnails:** Generate 200x200 thumbnails for previews
- **Storage:** Local filesystem for MVP, S3-compatible for production
- **Cleanup:** Automated script to remove media from deleted messages (30-day retention)

### 4. API Optimizations
- **Rate Limiting:** 100 req/min per user, 1000 req/min for premium
- **Pagination:** Default limit=50, max=100
- **Response Compression:** Gzip/Brotli via Caddy
- **Keep-Alive:** HTTP/2 for persistent connections

### 5. Memory Management
- **Go Garbage Collection:** GOGC=100 (default), tune based on monitoring
- **Connection Limits:** Max 500 WebSocket connections per instance
- **Buffer Sizes:** 4KB read/write buffers for WebSocket
- **Goroutine Pooling:** Limit concurrent goroutines to prevent memory spikes

---

## Deployment Plan

### Initial Setup (Ubuntu 24.04)

#### Prerequisites
```bash
# Run automated setup script
bash <(curl -fsSL https://raw.githubusercontent.com/your-repo/deploy/setup.sh)
```

#### Manual Steps (if needed)
1. **Update System**
   ```bash
   sudo apt update && sudo apt upgrade -y
   ```

2. **Install Docker**
   ```bash
   sudo apt install docker.io docker-compose-v2 -y
   sudo systemctl enable docker
   sudo usermod -aG docker $USER
   ```

3. **Clone Repository**
   ```bash
   git clone https://github.com/your-repo/messenger.git
   cd messenger
   ```

4. **Configure Environment**
   ```bash
   cp deploy/.env.example deploy/.env
   # Edit .env with real values
   nano deploy/.env
   ```

5. **Initialize Database**
   ```bash
   docker-compose -f deploy/docker-compose.yml up -d postgres
   sleep 10
   docker exec messenger-postgres psql -U messenger -d messenger -f /schema.sql
   ```

6. **Start All Services**
   ```bash
   docker-compose -f deploy/docker-compose.yml up -d
   ```

7. **Verify Deployment**
   ```bash
   curl http://localhost:8080/health
   # Expected: {"status": "ok"}
   ```

### Systemd Integration (Optional)
```bash
sudo cp deploy/systemd/* /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable messenger.target
sudo systemctl start messenger.target
```

### Firewall Configuration
```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 3478:3479/tcp
sudo ufw allow 3478:3479/udp
sudo ufw enable
```

---

## Security Considerations

### MVP Security
1. **Transport Security:** HTTPS via Caddy (Let's Encrypt)
2. **Password Security:** bcrypt hashing (cost=12)
3. **JWT Security:** HS256, 24-hour expiration, refresh tokens
4. **Input Validation:** Sanitize all user inputs
5. **Rate Limiting:** Prevent brute force attacks
6. **CORS:** Restrict to known origins

### Future Security (Stage 3-4)
1. **E2E Encryption:** Signal Protocol or Matrix Olm
2. **Key Management:** Client-side key generation, server only stores encrypted data
3. **Perfect Forward Secrecy:** Rotate encryption keys per session
4. **2FA:** TOTP-based two-factor authentication

---

## Monitoring & Logging

### Metrics to Monitor
- CPU usage per container
- Memory usage per container
- Database connection count
- Redis memory usage
- WebSocket connection count
- API response times (p50, p95, p99)
- Message delivery latency

### Logging Strategy
- **Format:** JSON structured logs
- **Destination:** stdout (captured by Docker)
- **Levels:** DEBUG (dev), INFO (prod), ERROR (always)
- **Retention:** 7 days for MVP, 30 days for production

### Health Checks
All services expose health endpoints:
- API: `GET /health`
- PostgreSQL: `pg_isready`
- Redis: `PING`
- Caddy: Built-in health checks

---

## MVP Payment System (Stub Implementation)

### Current Implementation
The MVP uses a **stub payment system** with the following behavior:

1. **No Real Payments:** No integration with Stripe, Yookassa, or any payment gateway
2. **Logging Only:** All "payments" are logged to `payment_logs` table with status `completed_stub`
3. **Instant Activation:** `user.is_premium` is set to `true` immediately upon "purchase"
4. **No Card Validation:** No credit card information is collected or validated

### API Behavior (MVP)
```json
POST /subscriptions/purchase
{
  "subscription_type": "premium_monthly"
}

Response:
{
  "success": true,
  "message": "MVP: Stub payment activated. No real charge applied.",
  "subscription": {
    "type": "premium_monthly",
    "end_date": "2024-02-15"
  }
}
```

### Future Integration (Stage 2)
- **Stripe:** For international payments (credit cards, Apple Pay, Google Pay)
- **Yookassa:** For Russian market
- **Webhooks:** Handle payment success/failure asynchronously
- **Refunds:** Automated refund logic
- **Invoicing:** Generate PDF invoices

---

## E2E Encryption Plan (Stage 3-4)

### Current State (MVP)
- **Transport Security Only:** HTTPS for all communications
- **Server-Side Storage:** Messages stored in plaintext in PostgreSQL
- **Server Access:** Server can read all messages

### Future Implementation
1. **Protocol:** Signal Protocol or Matrix Olm
2. **Key Exchange:** X3DH (Extended Triple Diffie-Hellman)
3. **Ratcheting:** Double Ratchet for forward secrecy
4. **Device Keys:** Each device has unique identity keys
5. **Group Encryption:** Sender Keys for efficient group messaging

### Migration Strategy
- Opt-in feature initially
- Gradual rollout to all users
- Compatibility layer for legacy clients
- Server stores encrypted blobs only

---

## Disaster Recovery

### Backup Strategy
1. **Database:** Daily full backups via `pg_dump`
2. **Media Files:** Weekly backups to S3-compatible storage
3. **Configuration:** Git-tracked, auto-deployed
4. **Retention:** 30-day retention for all backups

### Recovery Procedures
1. **Database Restoration:**
   ```bash
   docker exec -i messenger-postgres psql -U messenger < backup.sql
   ```
2. **Media Restoration:**
   ```bash
   rsync -av s3://backups/media/ /data/media/
   ```

---

## Performance Benchmarks

### Target Metrics (MVP on Target Hardware)
- **API Response Time:** < 100ms (p95)
- **WebSocket Latency:** < 50ms
- **Database Queries:** < 20ms (p95)
- **Concurrent Users:** 50-100
- **Messages per Second:** 500+

### Load Testing Plan
```bash
# Use k6 or Apache Bench
k6 run --vus 50 --duration 5m load-test.js
```

---

## Appendix

### Glossary
- **DM:** Direct Message (1-on-1 chat)
- **Group:** Multi-user chat (up to 200 members for free, unlimited for premium)
- **Channel:** One-to-many broadcast (owner posts, subscribers read)
- **Premium:** Paid subscription tier with enhanced features
- **TURN:** Traversal Using Relays around NAT (WebRTC relay server)

### References
- Go Fiber Documentation: https://docs.gofiber.io
- PostgreSQL Performance Tuning: https://wiki.postgresql.org/wiki/Tuning_Your_PostgreSQL_Server
- Redis Best Practices: https://redis.io/docs/management/optimization/
- Signal Protocol: https://signal.org/docs/

---

**Document Version:** 1.0.0  
**Last Updated:** 2024-01-15  
**Next Review:** After MVP deployment
