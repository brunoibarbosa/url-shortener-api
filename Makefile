include ./cmd/url-shortener/.env
export

MIGRATIONS_DIR=./internal/infra/database/pg/migrations
DB_DSN=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" status