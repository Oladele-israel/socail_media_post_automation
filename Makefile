# Makefile

# ─────────────────────────────────────────
# Load .env and export all variables
# so every make command can read them
# ─────────────────────────────────────────
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

# Build the DB URL from .env variables
# single source of truth — never hardcode credentials
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# ─────────────────────────────────────────
# APPLICATION
# ─────────────────────────────────────────

## run: start the application
run:
	@echo "🚀 Starting server..."
	@go run cmd/api/main.go

## build: compile binary into ./bin/api
build:
	@echo "🔨 Building binary..."
	@mkdir -p bin
	@go build -o bin/api cmd/api/main.go
	@echo "✅ Binary built at ./bin/api"

## start: run the compiled binary (run build first)
start: build
	@./bin/api

## tidy: clean up and download dependencies
tidy:
	@echo "📦 Tidying dependencies..."
	@go mod tidy
	@echo "✅ Done"

## test: run all tests
test:
	@echo "🧪 Running tests..."
	@go test ./... -v

## lint: check code for issues (requires golangci-lint)
lint:
	@golangci-lint run ./...

# ─────────────────────────────────────────
# DATABASE MIGRATIONS
# ─────────────────────────────────────────

## migrate-up: apply all pending migrations
migrate-up:
	@echo "⬆️  Running migrations up..."
	@migrate -path migrations -database "$(DB_URL)" up
	@echo "✅ Migrations applied"

## migrate-down: rollback the last migration
migrate-down:
	@echo "⬇️  Rolling back last migration..."
	@migrate -path migrations -database "$(DB_URL)" down 1
	@echo "✅ Rolled back"

## migrate-down-all: rollback ALL migrations — destroys all tables
migrate-down-all:
	@echo "⚠️  Rolling back ALL migrations..."
	@migrate -path migrations -database "$(DB_URL)" down -all
	@echo "✅ All migrations rolled back"

## migrate-version: check which migration version the DB is on
migrate-version:
	@migrate -path migrations -database "$(DB_URL)" version

## migrate-force: force-set migration version (fixes dirty state)
## Usage: make migrate-force VERSION=1
migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Provide a version: make migrate-force VERSION=1"; exit 1; \
	fi
	@migrate -path migrations -database "$(DB_URL)" force $(VERSION)
	@echo "✅ Forced to version $(VERSION)"

## migrate-create: create a new migration file pair
## Usage: make migrate-create NAME=create_posts_table
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Provide a name: make migrate-create NAME=create_posts_table"; exit 1; \
	fi
	@migrate create -ext sql -dir migrations -seq $(NAME)
	@echo "✅ Migration files created in migrations/"

# ─────────────────────────────────────────
# DATABASE UTILS
# ─────────────────────────────────────────

## db-connect: open a psql session to your database
db-connect:
	@psql -U $(DB_USER) -d $(DB_NAME) -h $(DB_HOST) -p $(DB_PORT)

## db-tables: list all tables in the database
db-tables:
	@psql -U $(DB_USER) -d $(DB_NAME) -h $(DB_HOST) -p $(DB_PORT) -c "\dt"

## db-create: create the database if it doesn't exist
db-create:
	@echo "🗄️  Creating database $(DB_NAME)..."
	@psql -U $(DB_USER) -h $(DB_HOST) -p $(DB_PORT) \
		-c "CREATE DATABASE $(DB_NAME);" 2>/dev/null || \
		echo "⚠️  Database already exists"

## db-drop: drop the database entirely — careful!
db-drop:
	@echo "💥 Dropping database $(DB_NAME)..."
	@psql -U $(DB_USER) -h $(DB_HOST) -p $(DB_PORT) \
		-c "DROP DATABASE IF EXISTS $(DB_NAME);"
	@echo "✅ Database dropped"

## db-reset: drop, recreate and re-run all migrations
db-reset: db-drop db-create migrate-up
	@echo "✅ Database reset complete"

# ─────────────────────────────────────────
# REDIS
# ─────────────────────────────────────────

## redis-connect: open a Redis CLI session
redis-connect:
	@redis-cli -h $(REDIS_HOST) -p $(REDIS_PORT)

## redis-keys: list all keys in Redis
redis-keys:
	@redis-cli -h $(REDIS_HOST) -p $(REDIS_PORT) KEYS "*"

## redis-flush: clear ALL Redis data — careful!
redis-flush:
	@echo "⚠️  Flushing all Redis data..."
	@redis-cli -h $(REDIS_HOST) -p $(REDIS_PORT) FLUSHALL
	@echo "✅ Redis flushed"

# ─────────────────────────────────────────
# HELP
# ─────────────────────────────────────────

## help: list all available commands
help:
	@echo ""
	@echo "Usage: make <command>"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/  /'
	@echo ""

.PHONY: run build start tidy test lint \
        migrate-up migrate-down migrate-down-all \
        migrate-version migrate-force migrate-create \
        db-connect db-tables db-create db-drop db-reset \
        redis-connect redis-keys redis-flush \
        help