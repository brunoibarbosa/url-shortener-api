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

.PHONY: setup-hooks mocks test

setup-hooks:
	@chmod +x scripts/setup-hooks.sh
	@./scripts/setup-hooks.sh

mocks:
	@echo "Generating mocks..."
	@mockgen -source=internal/domain/url/repository.go -destination=internal/mocks/url_repository_mock.go -package=mocks
	@mockgen -source=internal/domain/url/encrypter.go -destination=internal/mocks/url_encrypter_mock.go -package=mocks
	@mockgen -source=internal/domain/url/shortcode.go -destination=internal/mocks/shortcode_generator_mock.go -package=mocks
	@mockgen -source=internal/domain/user/repository.go -destination=internal/mocks/user_repository_mock.go -package=mocks
	@mockgen -source=internal/domain/user/encrypter.go -destination=internal/mocks/user_encrypter_mock.go -package=mocks
	@mockgen -source=internal/domain/session/repository.go -destination=internal/mocks/session_repository_mock.go -package=mocks
	@mockgen -source=internal/domain/session/encrypter.go -destination=internal/mocks/session_encrypter_mock.go -package=mocks
	@mockgen -source=internal/domain/session/service.go -destination=internal/mocks/token_service_mock.go -package=mocks
	@mockgen -source=internal/domain/session/state.go -destination=internal/mocks/state_service_mock.go -package=mocks
	@mockgen -source=internal/domain/bd/tx_manager.go -destination=internal/mocks/tx_manager_mock.go -package=mocks
	@echo "Mocks generated successfully!"

test:
	@echo "Running tests..."
	@go test ./... -v -cover
