# Contributing to Telegram Clone

Thank you for your interest in contributing to this project! This document provides guidelines and instructions for contributing.

## Table of Contents
- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Coding Standards](#coding-standards)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Testing](#testing)
- [Documentation](#documentation)

---

## Code of Conduct

### Our Standards
- Be respectful and inclusive
- Accept constructive criticism gracefully
- Focus on what is best for the community
- Show empathy towards other community members

### Unacceptable Behavior
- Harassment, trolling, or discriminatory comments
- Publishing others' private information
- Other conduct which could reasonably be considered inappropriate

---

## Getting Started

### Prerequisites
- Go 1.21 or higher
- PostgreSQL 16
- Redis 7
- Docker & Docker Compose
- Git

### Fork and Clone
```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/messenger.git
cd messenger

# Add upstream remote
git remote add upstream https://github.com/original-repo/messenger.git
```

---

## Development Setup

### Local Development

#### 1. Install Dependencies
```bash
cd backend
go mod download
```

#### 2. Setup Database
```bash
# Start PostgreSQL and Redis with Docker
docker compose -f deploy/docker-compose.yml up -d postgres redis

# Wait for services to be ready
sleep 5

# Run migrations
docker exec -it messenger-postgres psql -U messenger -d messenger -f /docker-entrypoint-initdb.d/schema.sql
```

#### 3. Configure Environment
```bash
cp backend/.env.example backend/.env
# Edit .env with your local configuration
```

#### 4. Run the Application
```bash
cd backend
go run cmd/api/main.go
```

The API should now be running on `http://localhost:8080`

### Using Docker (Recommended)
```bash
cd deploy
docker compose up -d
docker compose logs -f api
```

---

## Coding Standards

### Go Style Guide
Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

#### Key Points
- Use `gofmt` to format code
- Follow effective Go conventions
- Use meaningful variable names
- Keep functions small and focused
- Write comments for exported functions
- Handle errors explicitly

#### Example
```go
// Good
func (h *MessageHandler) SendMessage(c fiber.Ctx) error {
    userID := c.Locals("userID").(string)
    
    uid, err := uuid.Parse(userID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid user ID",
        })
    }
    
    // ... rest of implementation
}

// Bad
func (h *MessageHandler) SendMessage(c fiber.Ctx) error {
    u := c.Locals("userID").(string)
    id, _ := uuid.Parse(u) // Don't ignore errors!
    // ...
}
```

### File Organization
```
backend/
├── cmd/
│   └── api/
│       └── main.go           # Entry point
├── internal/
│   ├── handlers/             # HTTP handlers
│   └── models/               # Data models
├── pkg/
│   ├── auth/                 # Authentication utilities
│   ├── cache/                # Redis utilities
│   └── database/             # Database utilities
```

### Naming Conventions
- **Packages:** lowercase, single word (e.g., `auth`, `models`)
- **Files:** lowercase with underscores (e.g., `user_handler.go`)
- **Functions:** camelCase for private, PascalCase for exported
- **Constants:** PascalCase or ALL_CAPS
- **Variables:** camelCase

---

## Commit Guidelines

### Commit Message Format
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types
- **feat:** New feature
- **fix:** Bug fix
- **docs:** Documentation changes
- **style:** Code style changes (formatting, etc.)
- **refactor:** Code refactoring
- **test:** Adding or updating tests
- **chore:** Maintenance tasks

### Examples
```
feat(auth): add JWT refresh token endpoint

Implement token refresh functionality to allow users to
obtain new access tokens without re-authenticating.

Closes #123
```

```
fix(messages): correct WebSocket message broadcast

Fixed issue where messages were not being delivered to
all chat members due to incorrect channel subscription.

Fixes #456
```

### Commit Best Practices
- Keep commits atomic (one logical change per commit)
- Write clear, descriptive commit messages
- Reference issue numbers when applicable
- Keep commit history clean (use rebase when needed)

---

## Pull Request Process

### Before Submitting
1. **Update your fork**
   ```bash
   git fetch upstream
   git checkout main
   git merge upstream/main
   ```

2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**
   - Follow coding standards
   - Add tests for new features
   - Update documentation

4. **Run tests**
   ```bash
   go test ./...
   ```

5. **Format code**
   ```bash
   gofmt -w .
   ```

### Pull Request Template
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] No new warnings generated
```

### Review Process
1. Submit PR with clear description
2. Address reviewer feedback
3. Keep PR up to date with main branch
4. Wait for approval from maintainers
5. Squash commits if requested

---

## Testing

### Unit Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/handlers
```

### Integration Tests
```bash
# Start test environment
docker compose -f deploy/docker-compose.test.yml up -d

# Run integration tests
go test -tags=integration ./...
```

### Writing Tests
```go
func TestMessageHandler_SendMessage(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    handler := NewMessageHandler(db, nil)
    
    // Test cases
    tests := []struct {
        name    string
        userID  string
        wantErr bool
    }{
        {"valid message", "uuid", false},
        {"invalid user", "invalid", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

---

## Documentation

### Code Documentation
- Add godoc comments for all exported functions
- Include examples when appropriate
- Keep comments up to date with code changes

```go
// SendMessage handles sending a new message to a chat.
// It validates user membership, creates the message in the database,
// and broadcasts it via Redis pub/sub for real-time delivery.
//
// Example:
//   POST /api/v1/messages
//   {
//     "chat_id": "uuid",
//     "content": "Hello"
//   }
func (h *MessageHandler) SendMessage(c fiber.Ctx) error {
    // Implementation
}
```

### README Updates
Update README.md when:
- Adding new features
- Changing configuration
- Updating dependencies
- Modifying setup process

### API Documentation
Update `docs/TDD.md` with:
- New endpoints
- Request/response examples
- Error codes
- Authentication requirements

---

## Development Workflow

### 1. Pick an Issue
- Check open issues on GitHub
- Comment on issue to claim it
- Discuss approach if needed

### 2. Develop
- Create feature branch
- Write code following standards
- Add tests
- Update documentation

### 3. Test
- Run unit tests
- Perform manual testing
- Check for edge cases

### 4. Submit PR
- Push to your fork
- Create pull request
- Wait for review

### 5. Address Feedback
- Make requested changes
- Push updates to PR branch
- Re-request review

---

## Project Structure

### Backend Structure
```
backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── handlers/                # Request handlers
│   │   ├── auth.go
│   │   ├── messages.go
│   │   ├── chats.go
│   │   └── websocket.go
│   └── models/                  # Data models
│       ├── user.go
│       ├── message.go
│       └── chat.go
├── pkg/
│   ├── auth/                    # Authentication
│   ├── cache/                   # Redis client
│   └── database/                # Database client
├── go.mod
└── go.sum
```

### Adding New Features

#### 1. Add Model (if needed)
```go
// internal/models/feature.go
type Feature struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
    Name      string    `gorm:"type:varchar(255)" json:"name"`
    CreatedAt time.Time `json:"created_at"`
}
```

#### 2. Add Handler
```go
// internal/handlers/feature.go
type FeatureHandler struct {
    db *gorm.DB
}

func NewFeatureHandler(db *gorm.DB) *FeatureHandler {
    return &FeatureHandler{db: db}
}

func (h *FeatureHandler) Create(c fiber.Ctx) error {
    // Implementation
}
```

#### 3. Register Routes
```go
// cmd/api/main.go
featureHandler := handlers.NewFeatureHandler(db)
api.Post("/features", auth.Protected(), featureHandler.Create)
```

#### 4. Update Database Schema
```bash
# Add migration to database/migration_vX.sql
# Or update database/schema.sql for new installations
```

#### 5. Update Documentation
- Add API docs to `docs/TDD.md`
- Update README.md if it's a user-facing feature
- Add examples to `docs/FEATURES.md`

### Working with New Features (v2)

#### Wiki Pages
- Location: `backend/internal/models/wiki.go`, `backend/internal/handlers/wiki.go`
- Supports: Markdown, hierarchical structure, custom ordering
- Testing: Create page, update content, verify tree structure

#### Code Snippets
- Location: `backend/internal/models/code.go`, `backend/internal/handlers/code.go`
- Supports: Multiple languages, syntax highlighting, message linking
- Testing: Create snippet, retrieve by chat, verify language filtering

#### Temporary Roles
- Location: `backend/internal/models/temp_roles.go`, `backend/internal/handlers/temp_roles.go`
- Supports: Custom permissions, expiration, permission checking
- Testing: Grant role, check permission, verify expiration

#### RSS Aggregator
- Location: `backend/internal/models/rss.go`, `backend/internal/handlers/rss.go`
- Supports: RSS 2.0/Atom parsing, auto-refresh, duplicate detection
- Testing: Add feed, refresh feed, retrieve items

### Resource Optimization Guidelines

When adding new features, keep resource limits in mind:
- **Memory**: Total application memory should not exceed 900MB
- **Database**: Efficient queries with proper indexes
- **Cache**: Use Redis for frequently accessed data
- **Connections**: Limit database connections (max 20)

### Performance Testing

Before merging features:
```bash
# Run load test with Artillery
artillery quick --count 100 --num 10 http://localhost:8080/api/v1/messages

# Monitor resource usage
docker stats

# Check database query performance
docker exec messenger-postgres psql -U messenger -d messenger -c "EXPLAIN ANALYZE SELECT ..."
```

---

## Common Issues

### Database Connection Failed
```bash
# Check PostgreSQL is running
docker compose ps postgres

# Check connection settings in .env
DB_HOST=localhost
DB_PORT=5432
```

### Redis Connection Failed
```bash
# Check Redis is running
docker compose ps redis

# Test connection
docker exec -it messenger-redis redis-cli ping
```

### Port Already in Use
```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>
```

---

## Getting Help

### Resources
- [Technical Design Document](docs/TDD.md)
- [Roadmap](docs/ROADMAP.md)
- [Go Documentation](https://golang.org/doc/)
- [Fiber Documentation](https://docs.gofiber.io)

### Contact
- Open an issue for bugs or feature requests
- Discuss in GitHub Discussions
- Email: maintainer@example.com (if applicable)

---

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing! 🎉
