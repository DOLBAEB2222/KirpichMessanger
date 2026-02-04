# Changelog

All notable changes to KirpichMessenger will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2024-01-15

### Added
- **Wiki in Channels**: Create and manage documentation pages within channels
  - Hierarchical page structure with parent-child relationships
  - Markdown support
  - Draft and published states
  - Custom ordering
  - Slug-based URLs
  - Tree view API endpoint

- **Code Snippets**: Mini code editor with syntax highlighting
  - Support for 17+ programming languages
  - Message linking
  - Optional file names
  - Search by chat and language
  - Version tracking

- **Temporary Roles**: Time-limited role assignments
  - Grant roles for specific duration
  - Custom permissions
  - Target chats or channels
  - Automatic expiration
  - Permission checking API
  - Role templates (moderator, admin, custom)

- **RSS Aggregator**: Subscribe to RSS feeds in channels
  - RSS 2.0 and Atom support
  - Manual feed refresh
  - Duplicate detection
  - HTML sanitization
  - Feed metadata tracking
  - Error tracking and reporting

- **New API Endpoints**:
  - Wiki: 6 endpoints (create, get, update, delete, list, tree)
  - Code: 5 endpoints (create, get, update, delete, list)
  - Temp Roles: 7 endpoints (grant, get, update, revoke, list, list by user, check)
  - RSS: 6 endpoints (add, get, update, delete, list, refresh, get items)

- **Message Type**: Added `code` message type for code snippets

### Changed
- **Resource Optimization**: Optimized for Intel i3-2120, 4GB RAM
  - PostgreSQL: 800MB → 700MB
  - Redis: 300MB → 250MB
  - API: 1GB → 900MB
  - Caddy: 200MB → 150MB
  - coturn: 200MB → 150MB
  - Total memory usage: ~2.15GB (54% of system RAM)

- **Database Configuration**:
  - Reduced max connections: 100 → 80
  - Optimized shared buffers for 700MB limit
  - Reduced worker processes: 4 → 2
  - Tuned WAL settings for low I/O
  - Better caching strategy

- **Redis Configuration**:
  - Reduced memory limit: 256MB → 200MB
  - Optimized save intervals for less disk I/O
  - Connection pooling improvements

- **Application Configuration**:
  - Added GOMAXPROCS=2 for CPU limiting
  - Database connection pooling: max 20, idle 10
  - Connection lifetime: 5 minutes
  - Rate limiting improvements

### Database Schema
- **New Tables**:
  - `wiki_pages` - Channel documentation
  - `code_snippets` - Code with syntax highlighting
  - `temp_roles` - Time-limited permissions
  - `rss_feeds` - RSS feed subscriptions
  - `rss_items` - Parsed feed items

- **New Types**:
  - `code_language` - Programming languages (17 options)
  - `temp_role_type` - Role types (moderator, admin, custom)
  - `temp_role_target` - Target types (chat, channel)

- **Updated Types**:
  - `message_type` - Added `code` option

### Documentation
- Added [FEATURES.md](docs/FEATURES.md) - Detailed feature documentation
- Added [DEPLOYMENT_GUIDE.md](docs/DEPLOYMENT_GUIDE.md) - Complete deployment guide
- Added [UPGRADE.md](docs/UPGRADE.md) - Upgrade from v1 to v2
- Added [docs/README.md](docs/README.md) - Documentation index
- Updated README.md with new features
- Updated CONTRIBUTING.md with v2 development guidelines

### Scripts
- Added `scripts/migrate.sh` - Automated database migration
- Added `scripts/health-check.sh` - System health monitoring

### Performance
- Optimized for 500+ concurrent users
- Improved cache hit rates
- Reduced database query times
- Lower memory footprint
- Better CPU utilization

### Security
- HTML sanitization for RSS feeds
- Permission checking for all new features
- Input validation improvements
- Role-based access control

## [1.0.0] - 2024-01-01

### Added
- User authentication with JWT
- Direct messages (DM) with automatic chat creation
- Group chats
- Channels for broadcasting
- Real-time messaging with WebSocket
- Media sharing (images, videos, audio, files)
- Message compression
- Read receipts
- Typing indicators
- Online status tracking
- Premium subscriptions (stub implementation)
- Voice and video calls with WebRTC
- User contacts
- Blocked users
- Search functionality
- Message history with pagination
- Auto-named chats for DMs
- Member management
- Channel subscriptions

### Technology
- Go 1.21+ with Fiber v3
- PostgreSQL 15+
- Redis 7
- Caddy 2 for reverse proxy
- coturn for WebRTC TURN server
- Docker & Docker Compose

### Documentation
- Basic API documentation
- Deployment guide
- Contributing guidelines
- Technical design document

## [Unreleased]

### Planned
- E2E encryption
- Real payment integration (Stripe/Yookassa)
- Desktop client (Tauri)
- Mobile apps
- Video file compression
- Advanced search with filters
- Message reactions
- Message pinning
- Scheduled messages
- Bot API
- Plugin system
- Multi-language support
- Dark mode
- Custom themes
