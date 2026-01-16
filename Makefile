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

.PHONY: setup-hooks mocks test clean-docker

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

# Fast tests - No Docker required (for pre-commit hooks)
test-fast:
	@echo "Running fast tests..."
	@go test ./internal/domain/... ./internal/app/... ./internal/infra/service/... \
		-count=1 -cover -short -timeout 30s
	@echo ""
	@echo "Fast tests completed!"

# Integration tests - Requires Docker (for manual testing and CI/CD)
test-integration:
	@echo "Running integration tests..."
	@echo "Checking Docker availability..."
	@docker info > /dev/null 2>&1 || (echo "Docker is not running. Please start Docker Desktop." && exit 1)
	@echo ""
	@echo "Testing Redis repositories..."
	@go test ./internal/infra/repository/redis/... -count=1 -cover -timeout 3m
	@echo ""
	@echo "Testing PostgreSQL repositories..."
	@go test ./internal/infra/repository/pg/session -count=1 -cover -timeout 3m
	@go test ./internal/infra/repository/pg/url -count=1 -cover -timeout 3m
	@go test ./internal/infra/repository/pg/user -count=1 -cover -timeout 3m
	@echo ""
	@echo "Integration tests completed!"

# All tests - Complete test suite (for CI/CD)
test-all: test-fast test-integration
	@echo ""
	@echo "All tests passed successfully!"

# Property-based tests - Generate and test with random data
test-property:
	@echo "Running property-based tests..."
	@go test ./internal/infra/repository/pg/user -run "PropertyBased" -count=1 -timeout 5m
	@echo "Property-based tests completed!"

# Edge case tests - Test boundary conditions and special cases
test-edge:
	@echo "Running edge case tests..."
	@go test ./internal/infra/repository/pg/session -run "EdgeCase" -count=1 -timeout 3m
	@go test ./internal/infra/repository/pg/url -run "EdgeCase" -count=1 -timeout 3m
	@go test ./internal/infra/repository/pg/user -run "EdgeCase" -count=1 -timeout 3m
	@echo "Edge case tests completed!"

# E2E tests - Test complete API workflows
test-e2e:
	@echo "Running E2E tests..."
	@go test ./test/e2e/... -count=1 -timeout 5m
	@echo "E2E tests completed!"

# Default test command (runs fast tests for quick feedback)
test: test-fast

# Coverage report - Generates HTML coverage report
coverage:
	@echo "Generating coverage report..."
	@docker info > /dev/null 2>&1 || (echo "Docker is not running. Please start Docker Desktop." && exit 1)
	@go test ./... -count=1 -coverprofile=coverage.out -timeout 10m
	@go tool cover -func=coverage.out | grep total
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Open coverage.html in your browser to view detailed coverage"

clean-docker:
	@echo "Cleaning testcontainers..."
	@docker ps -aq --filter "label=org.testcontainers=true" | xargs -r docker rm -f 2>/dev/null || true
	@docker volume prune -f
	@echo "Cleanup done!"
