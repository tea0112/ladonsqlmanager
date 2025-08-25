.PHONY: help up down restart logs clean build test

# Default target
help: ## Show this help message
	@echo "Ladon SQL Manager - Docker Compose Commands"
	@echo "============================================="
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

up: ## Start the PostgreSQL database and pgAdmin
	docker compose up -d
	@echo "✅ Services started successfully!"
	@echo "📊 PostgreSQL: localhost:5432"
	@echo "🌐 pgAdmin: http://localhost:8080"
	@echo "   Email: admin@ladon.local"
	@echo "   Password: admin123"

down: ## Stop all services
	docker-compose down
	@echo "✅ Services stopped successfully!"

restart: ## Restart all services
	docker-compose restart
	@echo "✅ Services restarted successfully!"

logs: ## Show logs from all services
	docker-compose logs -f

logs-postgres: ## Show PostgreSQL logs
	docker-compose logs -f postgres

logs-pgadmin: ## Show pgAdmin logs
	docker-compose logs -f pgadmin

clean: ## Remove all containers, networks, and volumes
	docker-compose down -v --remove-orphans
	@echo "✅ All containers, networks, and volumes removed!"

build: ## Build the services
	docker-compose build

status: ## Show status of all services
	docker-compose ps

shell-postgres: ## Open a shell in the PostgreSQL container
	docker-compose exec postgres psql -U ladon_user -d ladon_db

shell-pgadmin: ## Open a shell in the pgAdmin container
	docker-compose exec pgadmin sh

# Database management
init-db: ## Initialize the database schema using GORM migrations
	@echo "🗄️  Initializing database schema with GORM migrations..."
	@echo "💡 The database will be automatically initialized when you run your application"
	@echo "💡 Or you can run: go run main.go (which will call manager.Init())"
	@echo "✅ Database is ready for GORM migrations!"

migrate: ## Run GORM migrations manually
	@echo "🗄️  Running GORM migrations..."
	go run cmd/migrate/main.go -action=migrate
	@echo "✅ Migrations completed!"

drop-tables: ## Drop all tables using GORM
	@echo "🗄️  Dropping all tables..."
	go run cmd/migrate/main.go -action=drop
	@echo "✅ Tables dropped!"

reset-db: ## Reset the database (drop and recreate using GORM)
	@echo "🔄 Resetting database..."
	go run cmd/migrate/main.go -action=reset
	@echo "✅ Database reset complete!"

# Development helpers
dev: up ## Start development environment
	@echo "🚀 Development environment ready!"
	@echo "💡 Run 'make test-app' to test the application"

test-app: ## Test the application with the database
	@echo "🧪 Testing application..."
	@echo "Make sure to set DB_STRING=postgres://ladon_user:ladon_password@localhost:5432/ladon_db?sslmode=disable"
	# cd example && go run main.go
	env DB_STRING='postgres://ladon_user:ladon_password@localhost:5432/ladon_db?sslmode=disable' go run ./example/main.go
stat main.go: no such file or directory

# Utility commands
health: ## Check health of all services
	docker-compose ps
	@echo ""
	@echo "🔍 Checking service health..."
	docker-compose exec postgres pg_isready -U ladon_user -d ladon_db

backup: ## Create a database backup
	@echo "💾 Creating database backup..."
	docker-compose exec postgres pg_dump -U ladon_user ladon_db > backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "✅ Backup created!"

restore: ## Restore database from backup (usage: make restore BACKUP_FILE=backup.sql)
	@if [ -z "$(BACKUP_FILE)" ]; then \
		echo "❌ Please specify backup file: make restore BACKUP_FILE=backup.sql"; \
		exit 1; \
	fi
	@echo "🔄 Restoring database from $(BACKUP_FILE)..."
	docker-compose exec -T postgres psql -U ladon_user -d ladon_db < $(BACKUP_FILE)
	@echo "✅ Database restored!"
