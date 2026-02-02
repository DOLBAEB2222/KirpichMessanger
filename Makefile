.PHONY: help build run test clean migrate docker-up docker-down docker-logs

# Default target
help:
	@echo "Messenger Application - Makefile Commands"
	@echo ""
	@echo "Development:"
	@echo "  make build         - Build the Go application"
	@echo "  make run           - Run the application locally"
	@echo "  make test          - Run tests"
	@echo "  make test-coverage - Run tests with coverage"
	@echo "  make lint          - Run linter"
	@echo "  make format        - Format code"
	@echo ""
	@echo "Database:"
	@echo "  make migrate       - Run database migrations"
	@echo "  make db-shell      - Open PostgreSQL shell"
	@echo "  make redis-cli     - Open Redis CLI"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up     - Start all services"
	@echo "  make docker-down   - Stop all services"
	@echo "  make docker-logs   - View logs"
	@echo "  make docker-restart - Restart all services"
	@echo "  make docker-clean  - Remove volumes and clean up"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make cleanup-media - Run media cleanup script"

# Build
build:
	@echo "Building application..."
	cd backend && go build -o bin/api cmd/api/main.go
	@echo "Build complete: backend/bin/api"

# Run
run:
	@echo "Starting application..."
	cd backend && go run cmd/api/main.go

# Tests
test:
	@echo "Running tests..."
	cd backend && go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	cd backend && go test -v -cover -coverprofile=coverage.out ./...
	cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: backend/coverage.html"

# Linting
lint:
	@echo "Running linter..."
	cd backend && golangci-lint run ./...

# Formatting
format:
	@echo "Formatting code..."
	cd backend && gofmt -w .
	@echo "Code formatted"

# Database
migrate:
	@echo "Running database migrations..."
	docker exec -i messenger-postgres psql -U messenger -d messenger < database/schema.sql
	@echo "Migrations complete"

db-shell:
	@echo "Opening PostgreSQL shell..."
	docker exec -it messenger-postgres psql -U messenger -d messenger

redis-cli:
	@echo "Opening Redis CLI..."
	docker exec -it messenger-redis redis-cli

# Docker
docker-up:
	@echo "Starting all services..."
	cd deploy && docker compose up -d
	@echo "Services started. Use 'make docker-logs' to view logs"

docker-down:
	@echo "Stopping all services..."
	cd deploy && docker compose down
	@echo "Services stopped"

docker-logs:
	cd deploy && docker compose logs -f

docker-restart:
	@echo "Restarting all services..."
	cd deploy && docker compose restart
	@echo "Services restarted"

docker-clean:
	@echo "Cleaning up Docker resources..."
	cd deploy && docker compose down -v
	@echo "Volumes removed"

docker-rebuild:
	@echo "Rebuilding and starting services..."
	cd deploy && docker compose up -d --build
	@echo "Services rebuilt and started"

# Cleanup
clean:
	@echo "Cleaning build artifacts..."
	rm -f backend/bin/api
	rm -f backend/coverage.out
	rm -f backend/coverage.html
	@echo "Clean complete"

cleanup-media:
	@echo "Running media cleanup script..."
	bash scripts/cleanup.sh
	@echo "Media cleanup complete"

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	cp backend/.env.example backend/.env
	cp deploy/.env.example deploy/.env
	@echo "Environment files created. Please update them with your configuration."
	@echo "Run 'make docker-up' to start services"

dev-reset:
	@echo "Resetting development environment..."
	make docker-down
	make docker-clean
	make docker-up
	sleep 10
	make migrate
	@echo "Development environment reset complete"

# Health checks
health:
	@echo "Checking service health..."
	@curl -s http://localhost:8080/health | jq . || echo "API not responding"
	@docker exec messenger-postgres pg_isready -U messenger || echo "PostgreSQL not ready"
	@docker exec messenger-redis redis-cli ping || echo "Redis not responding"

# Monitoring
stats:
	@echo "Service statistics:"
	@docker stats --no-stream messenger-api messenger-postgres messenger-redis messenger-caddy

# Backup
backup-db:
	@echo "Backing up database..."
	@mkdir -p backups
	@docker exec messenger-postgres pg_dump -U messenger messenger > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "Backup created in backups/"

# Production helpers
prod-deploy:
	@echo "Deploying to production..."
	@echo "WARNING: This is a production deployment!"
	@read -p "Continue? (y/N) " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		cd deploy && docker compose -f docker-compose.yml up -d --build; \
		echo "Production deployment complete"; \
	fi

prod-logs:
	cd deploy && docker compose logs -f --tail=100

prod-status:
	@echo "Production service status:"
	cd deploy && docker compose ps
