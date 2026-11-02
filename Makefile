# Migration settings
MIGRATIONS_DIR := migrations
STEPS ?= 1

# Load environment variables from .env
-include .env
export DATABASE_URL

.PHONY: migrate-create migrate-up migrate-down migrate-down-all migrate-version

# Create a new sequential migration.
# Usage: make migrate-create name=create_students
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=migration_name"; \
		exit 1; \
	fi
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# Apply all available migrations.
migrate-up:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "Error: DATABASE_URL is not set"; \
		exit 1; \
	fi
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

# Roll back the last N migrations.
# Usage: make migrate-down STEPS=2
migrate-down:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "Error: DATABASE_URL is not set"; \
		exit 1; \
	fi
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down $(STEPS)

# Roll back all applied migrations.
migrate-down-all:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "Error: DATABASE_URL is not set"; \
		exit 1; \
	fi
	@printf "WARNING: This will roll back ALL applied migrations.\nAre you sure? [y/N]: "; \
	read -r ans; \
	case "$$ans" in \
		y|Y) migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down ;; \
		*) echo "Migration rollback canceled" ;; \
	esac

# Show current migration version.
migrate-version:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "Error: DATABASE_URL is not set"; \
		exit 1; \
	fi
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

bot-run:
	@go mod tidy && \
	go run cmd/bot/main.go