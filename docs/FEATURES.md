# Unique Features Documentation

This document describes the unique features added to KirpichMessenger.

## Wiki in Channels

### Overview
Wiki pages allow channels to have structured documentation with hierarchical organization. Each wiki page supports Markdown formatting and can have parent-child relationships.

### Features
- **Markdown Support**: Full Markdown syntax for formatting
- **Hierarchical Structure**: Pages can have parent-child relationships
- **Version Control**: Track creation and modification dates
- **Publishing Control**: Draft and published states
- **Custom Ordering**: Organize pages with custom order
- **Slug-based URLs**: Clean, human-readable URLs

### API Endpoints
- `POST /api/v1/wiki` - Create wiki page
- `GET /api/v1/wiki/:channelId/:slug` - Get wiki page
- `PATCH /api/v1/wiki/:channelId/:slug` - Update wiki page
- `DELETE /api/v1/wiki/:channelId/:slug` - Delete wiki page
- `GET /api/v1/wiki/:channelId` - List all wiki pages in a channel
- `GET /api/v1/wiki/:channelId/tree` - Get hierarchical tree structure

### Usage Example
```bash
# Create a wiki page
curl -X POST https://example.com/api/v1/wiki \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": "123e4567-e89b-12d3-a456-426614174000",
    "slug": "getting-started",
    "title": "Getting Started",
    "content": "# Welcome\n\nThis is the getting started guide..."
  }'

# Get wiki tree
curl https://example.com/api/v1/wiki/123e4567-e89b-12d3-a456-426614174000/tree \
  -H "Authorization: Bearer $TOKEN"
```

## Code Snippets

### Overview
Code snippets allow users to share code with syntax highlighting. Each snippet is attached to a message and supports multiple programming languages.

### Supported Languages
- JavaScript, TypeScript
- Python, Go, Java
- C, C++, Rust
- PHP, Ruby
- SQL
- HTML, CSS, Bash
- JSON, XML, Markdown
- Other

### Features
- **Syntax Highlighting**: Language-aware highlighting
- **File Names**: Optional file name for context
- **Message Linking**: Each snippet is linked to a message
- **Version Tracking**: Track modifications
- **Search by Language**: Filter snippets by programming language

### API Endpoints
- `POST /api/v1/code` - Create code snippet
- `GET /api/v1/code/:id` - Get code snippet
- `PATCH /api/v1/code/:id` - Update code snippet
- `DELETE /api/v1/code/:id` - Delete code snippet
- `GET /api/v1/code/chat/:chatId` - List code snippets in a chat
- `GET /api/v1/code/message/:messageId` - Get code snippet by message

### Usage Example
```bash
# Create a code snippet
curl -X POST https://example.com/api/v1/code \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "message_id": "123e4567-e89b-12d3-a456-426614174000",
    "chat_id": "123e4567-e89b-12d3-a456-426614174001",
    "language": "python",
    "code": "print(\"Hello, World!\")",
    "file_name": "hello.py"
  }'

# List all Python snippets in a chat
curl "https://example.com/api/v1/code/chat/123e4567-e89b-12d3-a456-426614174001?language=python" \
  -H "Authorization: Bearer $TOKEN"
```

## Temporary Roles

### Overview
Temporary roles allow administrators to grant time-limited permissions to users for specific chats or channels. This is useful for moderators, temporary admins, or special event roles.

### Features
- **Time-limited**: Roles expire automatically after a specified duration
- **Custom Permissions**: Define specific permissions for each role
- **Role Types**: Pre-defined types (moderator, admin) or custom roles
- **Target Types**: Apply to chats or channels
- **Permission Checking**: API to check if a user has specific permissions
- **Activity Tracking**: Track role grants and expirations

### Standard Permissions
- `manage_roles` - Grant/revoke roles
- `edit_messages` - Edit any message
- `delete_messages` - Delete any message
- `manage_members` - Add/remove members
- `manage_wiki` - Edit wiki pages
- `manage_rss` - Manage RSS feeds
- `admin` - All permissions

### API Endpoints
- `POST /api/v1/temp-roles` - Grant temporary role
- `GET /api/v1/temp-roles/:id` - Get role details
- `PATCH /api/v1/temp-roles/:id` - Update role (extend/disable)
- `DELETE /api/v1/temp-roles/:id` - Revoke role
- `GET /api/v1/temp-roles/target/:targetId/:targetType` - List roles for a target
- `GET /api/v1/temp-roles/user/:userId` - List roles for a user
- `GET /api/v1/temp-roles/check/:userId/:targetId` - Check user permission

### Usage Example
```bash
# Grant moderator role for 24 hours
curl -X POST https://example.com/api/v1/temp-roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_id": "123e4567-e89b-12d3-a456-426614174000",
    "target_type": "channel",
    "user_id": "123e4567-e89b-12d3-a456-426614174002",
    "role_type": "moderator",
    "permissions": ["edit_messages", "delete_messages"],
    "duration_hours": 24
  }'

# Check if user has edit permission
curl "https://example.com/api/v1/temp-roles/check/123e4567-e89b-12d3-a456-426614174002/123e4567-e89b-12d3-a456-426614174000?target_type=channel&permission=edit_messages" \
  -H "Authorization: Bearer $TOKEN"
```

## RSS Aggregator

### Overview
RSS aggregator allows channels to subscribe to RSS/Atom feeds. New items from feeds are automatically fetched and can be posted as messages to the channel.

### Features
- **RSS 2.0 & Atom Support**: Parse standard feed formats
- **Auto-refresh**: Manual or scheduled feed refresh
- **Content Processing**: Strip HTML tags from descriptions
- **Duplicate Detection**: GUID-based duplicate prevention
- **Feed Metadata**: Track feed title, description, icon
- **Error Tracking**: Log fetch errors for troubleshooting

### API Endpoints
- `POST /api/v1/rss` - Add RSS feed to channel
- `GET /api/v1/rss/:id` - Get RSS feed details
- `PATCH /api/v1/rss/:id` - Update RSS feed (URL, enable/disable)
- `DELETE /api/v1/rss/:id` - Remove RSS feed
- `GET /api/v1/rss` - List all RSS feeds for subscribed channels
- `GET /api/v1/rss/:id/items` - Get feed items (paginated)
- `POST /api/v1/rss/:id/refresh` - Manually refresh feed

### Usage Example
```bash
# Add RSS feed to channel
curl -X POST https://example.com/api/v1/rss \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": "123e4567-e89b-12d3-a456-426614174000",
    "url": "https://blog.example.com/feed.xml"
  }'

# Get feed items
curl "https://example.com/api/v1/rss/123e4567-e89b-12d3-a456-426614174003/items?limit=20" \
  -H "Authorization: Bearer $TOKEN"

# Refresh feed
curl -X POST https://example.com/api/v1/rss/123e4567-e89b-12d3-a456-426614174003/refresh \
  -H "Authorization: Bearer $TOKEN"
```

## Performance Optimizations

### Database Optimization
- **Reduced Memory Usage**: PostgreSQL configured for 700MB limit
- **Optimized Connections**: Max 80 connections
- **Efficient Work Processes**: 2 worker processes instead of 4
- **Reduced WAL Size**: Smaller WAL files for less I/O
- **Better Caching**: LRU cache policy for frequent queries

### Redis Optimization
- **Memory Limit**: 250MB with LRU eviction
- **Persistence**: Optimized save intervals
- **Connection Pooling**: Reduced overhead

### Application Optimization
- **GOMAXPROCS**: Limited to 2 CPU cores
- **Database Pool**: Max 20 connections, 10 idle
- **Connection Lifetime**: 5-minute connection reuse
- **Rate Limiting**: Built-in request throttling

### Total Resource Usage
- **Memory**: ~2.15GB (54% of 4GB system)
- **CPU**: ~4 cores (100% of i3-2120)
- **Disk**: 20GB SSD minimum recommended

## Security Considerations

### Wiki Pages
- Only creators can edit their pages
- Published/unpublished state control
- No HTML/Script injection (Markdown only)

### Code Snippets
- Sanitized code storage
- No code execution
- Read-only access for non-owners

### Temporary Roles
- Permission checks for all operations
- Auto-expiration prevents privilege escalation
- Audit trail for role grants

### RSS Feeds
- URL validation before fetch
- HTML sanitization
- Timeout limits (30 seconds)
- Content size limits

## Future Enhancements

### Wiki
- Rich text editor with WYSIWYG
- Image uploads in wiki pages
- Page history and versioning
- Collaboration features (comments)

### Code Snippets
- Code execution playground
- Diff viewer for changes
- Code formatting/linting
- Shareable snippet URLs

### Temporary Roles
- Role templates
- Bulk role assignment
- Role renewal notifications
- Role activity reports

### RSS Aggregator
- Scheduled auto-refresh (cron jobs)
- Full-text search in feed items
- Keyword filtering
- Multiple feeds per channel
