.PHONY: help install dev build test test-help test-setup test-teardown test-unit test-integration clean docker-up docker-down backend-dev frontend-dev migrate-up migrate-down migrate-create migrate-seed migrate-status site-freshness-seed site-freshness-generate test-security test-idor k8s-provision k8s-setup k8s-verify k8s-deploy-prod k8s-deploy-staging openapi-validate openapi-serve openapi-build deploy-vps deploy-vps-status deploy-vps-logs deploy-vps-down

# Compose project + network names stay in sync across targets
PROJECT_NAME := $(if $(COMPOSE_PROJECT_NAME),$(COMPOSE_PROJECT_NAME),$(notdir $(CURDIR)))
DEV_NETWORK := $(PROJECT_NAME)_default

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test-help: ## Show quick test commands reference
	@cd backend && bash test-commands.sh

install: ## Install all dependencies
	@echo "Installing backend dependencies..."
	cd backend && go mod download
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@echo "Installing mobile dependencies..."
	cd mobile && npm install
	@echo "✓ All dependencies installed"

build: ## Build backend, frontend, and mobile
	@echo "Building backend..."
	cd backend && go build -o bin/api ./cmd/api
	@echo "Building frontend..."
	cd frontend && npm run build
	@echo "Building mobile (iOS)..."
	cd mobile && npm run ios -- --configuration Release || echo "⚠ Mobile iOS build skipped (requires macOS)"
	@echo "✓ Build complete"

test-setup: ## Set up test environment (containers + migrations + env)
	@echo "Setting up test environment configuration..."
	@cd backend && bash setup-test-env.sh
	@echo "Starting test containers (Postgres + Redis + OpenSearch)..."
	docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for Redis on localhost:6380..."
	@bash -c 'for i in {1..60}; do if docker compose -f docker-compose.test.yml exec -T redis-test redis-cli ping >/dev/null 2>&1; then echo "Redis is ready"; exit 0; fi; sleep 1; done; echo "Redis failed to become ready"; exit 1'
	@echo "Waiting for test Postgres on localhost:5437..."
	@bash -c 'until pg_isready -h localhost -p 5437 -U clpr -d clpr_test >/dev/null 2>&1; do sleep 1; done'
	@echo "Postgres is ready. Running test migrations..."
	@if command -v migrate > /dev/null; then \
		migrate -path backend/migrations -database "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable" up || true; \
	else \
		echo "Warning: golang-migrate not installed. Skipping migrations."; \
		echo "Install with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
	fi
	@echo "Waiting for OpenSearch on localhost:9201..."
	@bash -c 'for i in {1..60}; do if curl -f -s http://localhost:9201/_cluster/health >/dev/null 2>&1; then echo "OpenSearch is ready"; exit 0; fi; sleep 1; done; echo "OpenSearch failed to become ready"; exit 1'
	@echo "Seeding OpenSearch with test indices..."
	@if [ -f scripts/test-seed-opensearch.sh ]; then \
		OPENSEARCH_URL=http://localhost:9201 bash scripts/test-seed-opensearch.sh; \
	else \
		echo "Warning: test-seed-opensearch.sh not found"; \
	fi
	@echo "Seeding E2E test data (clips, users, subscriptions)..."
	@if [ -f scripts/test-seed-e2e.sh ]; then \
		TEST_DATABASE_HOST=localhost \
		TEST_DATABASE_PORT=5437 \
		TEST_DATABASE_USER=clpr \
		TEST_DATABASE_PASSWORD=clpr_password \
		TEST_DATABASE_NAME=clpr_test \
		OPENSEARCH_URL=http://localhost:9201 \
		bash scripts/test-seed-e2e.sh; \
	else \
		echo "Warning: test-seed-e2e.sh not found"; \
	fi
	@echo "✓ Test environment ready"

test-teardown: ## Tear down test environment (containers)
	@echo "Stopping test containers..."
	docker compose -f docker-compose.test.yml down
	@echo "✓ Test environment teardown complete"

test: ## Run all tests (unit by default; set INTEGRATION=1 and/or E2E=1 to expand)
	@echo "Running backend tests with verbose output..."
	@cd backend && INTEGRATION=$(INTEGRATION) E2E=$(E2E) bash run-tests-verbose.sh
	@if [ "$(E2E)" = "1" ]; then \
		echo "Starting backend API for frontend E2E..."; \
		mkdir -p .tmp; \
		(bash -c '\
			cd backend && \
			set -a && source .env.test && set +a && \
			PORT=8080 \
			GIN_MODE=debug \
			BASE_URL=http://127.0.0.1:5173 \
			DB_HOST=localhost \
			DB_PORT=5437 \
			DB_USER=clpr \
			DB_PASSWORD=clpr_password \
			DB_NAME=clpr_test \
			REDIS_HOST=localhost \
			REDIS_PORT=6380 \
			OPENSEARCH_URL=http://localhost:9201 \
			CORS_ALLOWED_ORIGINS=http://127.0.0.1:5173 \
			RATE_LIMIT_WHITELIST_IPS=127.0.0.1 \
			FEATURE_ANALYTICS=false \
			go run ./cmd/api \
		') > .tmp/backend-e2e.log 2>&1 & echo $$! > .tmp/backend-e2e.pid; \
		echo "Backend started (PID: $$(cat .tmp/backend-e2e.pid))"; \
		sleep 5; \
		echo "Running frontend E2E tests..."; \
		cd frontend && npm run test:e2e; \
		echo "Stopping backend API..."; \
		if [ -f .tmp/backend-e2e.pid ]; then kill $$(cat .tmp/backend-e2e.pid) || true; rm -f .tmp/backend-e2e.pid; fi; \
	else \
		echo "Skipping frontend E2E tests (set E2E=1 to enable)"; \
	fi
	@if [ "$(INTEGRATION)" != "1" ] && [ "$(E2E)" != "1" ]; then \
		echo "Running mobile tests..."; \
		cd mobile && npm run test || echo "Mobile tests not configured"; \
	else \
		echo "Skipping mobile tests (not needed for integration/E2E)"; \
	fi
	@if [ "$(INTEGRATION)" = "1" ] || [ "$(E2E)" = "1" ]; then \
		$(MAKE) test-teardown; \
	fi
	@echo "✓ Tests complete"

test-unit: ## Run unit tests only (fast, verbose)
	@echo "Running backend unit tests..."
	cd backend && bash run-tests-verbose.sh
	@echo "Running frontend unit tests..."
	cd frontend && npm run test -- run
	@echo "✓ Unit tests complete"

test-integration: ## Run integration tests (requires Docker, verbose)
	@$(MAKE) test-setup
	@echo "Running backend integration tests..."
	cd backend && go test -v -tags=integration -race -parallel=4 ./tests/integration/...
	@$(MAKE) test-teardown
	@echo "✓ Integration tests complete"

test-integration-coverage: ## Run integration tests with coverage report
	@echo "Starting test database..."
	docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Running database migrations..."
	migrate -path backend/migrations -database "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable" up || true
	@echo "Running integration tests with coverage..."
	cd backend && go test -v -tags=integration -race -parallel=4 -coverprofile=coverage-integration.out -covermode=atomic ./tests/integration/...
	@echo "Generating coverage report..."
	cd backend && go tool cover -html=coverage-integration.out -o coverage-integration.html
	@echo "Calculating coverage percentage..."
	@cd backend && go tool cover -func=coverage-integration.out | grep total | awk '{print "Integration test coverage: " $$3}'
	@echo "Coverage report generated at backend/coverage-integration.html"
	@echo "Stopping test database..."
	docker compose -f docker-compose.test.yml down
	@echo "✓ Integration tests with coverage complete"

test-integration-auth: ## Run authentication integration tests only
	@echo "Starting test database..."
	docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Running database migrations..."
	migrate -path backend/migrations -database "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable" up || true
	@echo "Running authentication integration tests..."
	cd backend && go test -v -tags=integration ./tests/integration/auth/...
	@echo "Stopping test database..."
	docker compose -f docker-compose.test.yml down
	@echo "✓ Authentication tests complete"

test-integration-submissions: ## Run submission integration tests only
	@echo "Starting test database..."
	docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Running database migrations..."
	migrate -path backend/migrations -database "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable" up || true
	@echo "Running submission integration tests..."
	cd backend && go test -v -tags=integration ./tests/integration/submissions/...
	@echo "Stopping test database..."
	docker compose -f docker-compose.test.yml down
	@echo "✓ Submission tests complete"

test-integration-engagement: ## Run engagement integration tests only
	@echo "Starting test database..."
	docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Running database migrations..."
	migrate -path backend/migrations -database "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable" up || true
	@echo "Running engagement integration tests..."
	cd backend && go test -v -tags=integration ./tests/integration/engagement/...
	@echo "Stopping test database..."
	docker compose -f docker-compose.test.yml down
	@echo "✓ Engagement tests complete"

test-integration-premium: ## Run premium integration tests only
	@echo "Starting test database..."
	docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Running database migrations..."
	migrate -path backend/migrations -database "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable" up || true
	@echo "Running premium integration tests..."
	cd backend && go test -v -tags=integration ./tests/integration/premium/...
	@echo "Stopping test database..."
	docker compose -f docker-compose.test.yml down
	@echo "✓ Premium tests complete"

test-integration-stripe: ## Run Stripe subscription & payment integration tests only
	@echo "Starting test database..."
	docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Running database migrations..."
	migrate -path backend/migrations -database "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable" up || true
	@echo "Running Stripe integration tests..."
	@echo "Note: Tests use Stripe test mode keys. Set TEST_STRIPE_SECRET_KEY and TEST_STRIPE_WEBHOOK_SECRET env vars for full testing."
	cd backend && go test -v -tags=integration ./tests/integration/premium/ -run "TestWebhook.*|TestEntitlement.*|TestProration.*|TestPaymentFailure.*"
	@echo "Stopping test database..."
	docker compose -f docker-compose.test.yml down
	@echo "✓ Stripe integration tests complete"

test-integration-search: ## Run search integration tests only
	@echo "Starting test database..."
	docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Running database migrations..."
	migrate -path backend/migrations -database "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable" up || true
	@echo "Running search integration tests..."
	cd backend && go test -v -tags=integration ./tests/integration/search/...
	@echo "Stopping test database..."
	docker compose -f docker-compose.test.yml down
	@echo "✓ Search tests complete"

test-integration-api: ## Run API integration tests only
	@echo "Starting test database..."
	docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Running database migrations..."
	migrate -path backend/migrations -database "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable" up || true
	@echo "Running API integration tests..."
	cd backend && go test -v -tags=integration ./tests/integration/api/...
	@echo "Stopping test database..."
	docker compose -f docker-compose.test.yml down
	@echo "✓ API tests complete"

test-integration-clips: ## Run clip management integration tests only
	@echo "Starting test database..."
	docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Running database migrations..."
	migrate -path backend/migrations -database "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable" up || true
	@echo "Running clip integration tests..."
	cd backend && go test -v -tags=integration ./tests/integration/clips/...
	@echo "Stopping test database..."
	docker compose -f docker-compose.test.yml down
	@echo "✓ Clip tests complete"

test-e2e: ## Run frontend E2E tests
	@echo "Running frontend E2E tests..."
	cd frontend && npm run test:e2e
	@echo "✓ E2E tests complete"

test-frontend: ## Run frontend unit tests only (verbose)
	@echo "Running frontend unit tests..."
	cd frontend && npm run test -- run
	@echo "✓ Frontend unit tests complete"

test-frontend-headed: ## Run frontend unit tests in headed mode (UI visible)
	@echo "Running frontend unit tests in headed mode..."
	cd frontend && npm run test
	@echo "✓ Frontend unit tests complete"

test-frontend-ui: ## Run frontend tests with Vitest UI
	@echo "Opening Vitest UI dashboard..."
	cd frontend && npm run test:ui
	@echo "✓ Vitest UI dashboard opened"

test-frontend-e2e: ## Run frontend Playwright E2E tests (verbose)
	@echo "Checking backend API availability..."
	@if curl -s http://localhost:8080/health > /dev/null 2>&1; then \
		echo "✓ Backend API is running"; \
	else \
		echo "⚠ Backend API not found at http://localhost:8080"; \
		echo "  Note: Backend E2E tests require the API to be running."; \
		echo "  Start with: make docker-dev-build && make docker-dev"; \
	fi
	@echo "Running Playwright E2E tests..."
	cd frontend && bash run-playwright-tests.sh
	@echo "✓ Frontend E2E tests complete"

test-frontend-e2e-ui: ## Run frontend Playwright E2E tests with UI (interactive)
	@echo "Opening Playwright UI..."
	cd frontend && npm run test:e2e:ui
	@echo "✓ Playwright UI closed"

test-frontend-e2e-report: ## View the last Playwright E2E test report
	@echo "Opening Playwright test report..."
	cd frontend && npx playwright show-report
	@echo "✓ Report viewer closed"

test-frontend-coverage: ## Run frontend unit tests with coverage report
	@echo "Running frontend tests with coverage..."
	cd frontend && npm run test:coverage
	@echo "✓ Frontend coverage tests complete"

test-frontend-all: ## Run all frontend tests (unit + E2E)
	@echo "Running all frontend tests..."
	@$(MAKE) test-frontend
	@$(MAKE) test-frontend-e2e
	@echo "✓ All frontend tests complete"

test-e2e-setup: ## Set up E2E test configurations (CDN failover, Stripe, etc.)
	@cd frontend && bash setup-e2e-tests.sh

test-frontend-help: ## Show frontend test command options
	@bash frontend/test-commands.sh

test-coverage: ## Run tests with coverage report
	@echo "Running backend tests with coverage..."
	cd backend && go test -coverprofile=coverage.out -covermode=atomic ./...
	@echo "Generating coverage report..."
	cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at backend/coverage.html"
	@echo "✓ Coverage tests complete"

test-load: ## Run all load tests (requires k6)
	@if command -v k6 > /dev/null; then \
		echo "Running all load tests..."; \
		k6 run backend/tests/load/scenarios/feed_browsing.js; \
		k6 run backend/tests/load/scenarios/clip_detail.js; \
		k6 run backend/tests/load/scenarios/search.js; \
		k6 run backend/tests/load/scenarios/comments.js; \
		k6 run backend/tests/load/scenarios/authentication.js; \
		k6 run backend/tests/load/scenarios/mixed_behavior.js; \
		echo "✓ All load tests complete"; \
	else \
		echo "Error: k6 is not installed"; \
		echo "Install it with: brew install k6 (macOS) or visit https://k6.io/docs/getting-started/installation/"; \
		exit 1; \
	fi

test-load-feed: ## Run feed browsing load test
	@k6 run backend/tests/load/scenarios/feed_browsing.js

test-load-clip: ## Run clip detail view load test
	@k6 run backend/tests/load/scenarios/clip_detail.js

test-load-search: ## Run search load test
	@k6 run backend/tests/load/scenarios/search.js

test-load-comments: ## Run comments load test
	@k6 run backend/tests/load/scenarios/comments.js

test-load-submit: ## Run submission load test (requires AUTH_TOKEN)
	@k6 run backend/tests/load/scenarios/submit.js

test-load-auth: ## Run authentication load test
	@k6 run backend/tests/load/scenarios/authentication.js

test-load-rate-limiting: ## Run rate limiting accuracy and performance test (requires AUTH_TOKEN)
	@k6 run backend/tests/load/scenarios/rate_limiting.js

test-load-report: ## Generate comprehensive load test report
	@if command -v k6 > /dev/null; then \
		echo "Generating comprehensive load test report..."; \
		cd backend/tests/load && ./generate_report.sh; \
	else \
		echo "Error: k6 is not installed"; \
		echo "Install it with: brew install k6 (macOS) or visit https://k6.io/docs/getting-started/installation/"; \
		exit 1; \
	fi

test-load-mixed: ## Run mixed user behavior load test
	@k6 run backend/tests/load/scenarios/mixed_behavior.js

test-load-moderation-ban-sync: ## Run moderation ban sync performance test
	@k6 run backend/tests/load/scenarios/moderation_ban_sync.js

test-load-moderation-audit-logs: ## Run moderation audit log query performance test
	@k6 run backend/tests/load/scenarios/moderation_audit_logs.js

test-load-moderation-permissions: ## Run moderation permission check performance test
	@k6 run backend/tests/load/scenarios/moderation_permissions.js

test-load-moderation-stress: ## Run comprehensive moderation stress test
	@k6 run backend/tests/load/scenarios/moderation_stress.js

test-load-moderation-all: ## Run all moderation performance tests
	@echo "Running all moderation performance tests..."
	@failed=0; \
	echo "=== Ban Sync Performance Test ==="; \
	k6 run backend/tests/load/scenarios/moderation_ban_sync.js || failed=$$((failed + 1)); \
	echo ""; \
	echo "=== Audit Log Query Performance Test ==="; \
	k6 run backend/tests/load/scenarios/moderation_audit_logs.js || failed=$$((failed + 1)); \
	echo ""; \
	echo "=== Permission Check Performance Test ==="; \
	k6 run backend/tests/load/scenarios/moderation_permissions.js || failed=$$((failed + 1)); \
	echo ""; \
	echo "=== Moderation Stress Test ==="; \
	k6 run backend/tests/load/scenarios/moderation_stress.js || failed=$$((failed + 1)); \
	echo ""; \
	if [ $$failed -eq 0 ]; then \
		echo "✓ All moderation performance tests passed"; \
	else \
		echo "✗ $$failed moderation test(s) failed"; \
		exit 1; \
	fi

test-load-baseline-capture: ## Capture performance baselines (requires VERSION env var)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION environment variable required in semantic versioning format"; \
		echo "Usage: make test-load-baseline-capture VERSION=vX.Y.Z"; \
		echo "Example: make test-load-baseline-capture VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "Capturing baseline for version $(VERSION)..."; \
	cd backend/tests/load && ./scripts/capture_baseline.sh $(VERSION)

test-load-baseline-compare: ## Compare against baseline (requires VERSION env var)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION environment variable required in semantic versioning format"; \
		echo "Usage: make test-load-baseline-compare VERSION=vX.Y.Z"; \
		echo "       make test-load-baseline-compare VERSION=current"; \
		echo "Example: make test-load-baseline-compare VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "Comparing against baseline $(VERSION)..."; \
	cd backend/tests/load && ./scripts/compare_baseline.sh $(VERSION)

test-load-html: ## Generate HTML reports for all load tests
	@if command -v k6 > /dev/null; then \
		echo "Generating HTML reports for all load tests..."; \
		cd backend/tests/load && ./scripts/generate_html_report.sh all; \
	else \
		echo "Error: k6 is not installed"; \
		echo "Install it with: brew install k6 (macOS) or visit https://k6.io/docs/getting-started/installation/"; \
		exit 1; \
	fi

# API Endpoint Benchmarks (Top 20 Endpoints with SLO Enforcement)

test-benchmark-feed-list: ## Run feed list endpoint benchmark (p50<20ms, p95<75ms)
	@if command -v k6 > /dev/null; then \
		echo "Running feed list endpoint benchmark..."; \
		k6 run backend/tests/load/scenarios/benchmarks/feed_list.js; \
	else \
		echo "Error: k6 is not installed"; \
		exit 1; \
	fi

test-benchmark-clip-detail: ## Run clip detail endpoint benchmark (p50<15ms, p95<50ms)
	@if command -v k6 > /dev/null; then \
		echo "Running clip detail endpoint benchmark..."; \
		k6 run backend/tests/load/scenarios/benchmarks/clip_detail.js; \
	else \
		echo "Error: k6 is not installed"; \
		exit 1; \
	fi

test-benchmark-search: ## Run search endpoint benchmark (p50<30ms, p95<100ms)
	@if command -v k6 > /dev/null; then \
		echo "Running search endpoint benchmark..."; \
		k6 run backend/tests/load/scenarios/benchmarks/search.js; \
	else \
		echo "Error: k6 is not installed"; \
		exit 1; \
	fi

test-benchmarks-all: ## Run all endpoint benchmarks
	@if command -v k6 > /dev/null; then \
		echo "Running all endpoint benchmarks..."; \
		cd backend/tests/load && ./run_all_benchmarks.sh; \
	else \
		echo "Error: k6 is not installed"; \
		exit 1; \
	fi

test-benchmarks-with-profiling: ## Run benchmarks with query profiling
	@if command -v k6 > /dev/null; then \
		echo "Running benchmarks with database query profiling..."; \
		for script in backend/tests/load/scenarios/benchmarks/*.js; do \
			endpoint=$$(basename "$$script" .js); \
			echo ""; \
			echo "Profiling endpoint: $$endpoint"; \
			cd backend/tests/load && ./profile_queries.sh "$$endpoint" 60 || true; \
		done; \
		echo "✓ All benchmarks with profiling complete"; \
		echo "View reports in backend/tests/load/profiles/benchmarks/ and profiles/queries/"; \
	else \
		echo "Error: k6 is not installed"; \
		exit 1; \
	fi

test-profile-queries: ## Profile queries for a specific endpoint (usage: make test-profile-queries ENDPOINT=feed_list DURATION=60)
	@if [ -z "$(ENDPOINT)" ]; then \
		echo "Error: ENDPOINT required"; \
		echo "Usage: make test-profile-queries ENDPOINT=feed_list DURATION=60"; \
		echo "Available endpoints: feed_list, clip_detail, search, etc."; \
		exit 1; \
	fi
	@DURATION=$${DURATION:-60}; \
	echo "Profiling endpoint $(ENDPOINT) for $$DURATION seconds..."; \
	cd backend/tests/load && ./profile_queries.sh $(ENDPOINT) $$DURATION

test-stress: ## Run stress test (push system beyond capacity)
	@if command -v k6 > /dev/null; then \
		echo "Running stress test (20 min full)..."; \
		k6 run backend/tests/load/scenarios/stress.js; \
	else \
		echo "Error: k6 is not installed"; \
		echo "Install it with: brew install k6 (macOS) or visit https://k6.io/docs/getting-started/installation/"; \
		exit 1; \
	fi

test-stress-lite: ## Run stress test lite version (5 min for CI)
	@if command -v k6 > /dev/null; then \
		echo "Running stress test lite (5 min)..."; \
		k6 run -e DURATION_MULTIPLIER=0.25 backend/tests/load/scenarios/stress.js; \
	else \
		echo "Error: k6 is not installed"; \
		echo "Install it with: brew install k6 (macOS) or visit https://k6.io/docs/getting-started/installation/"; \
		exit 1; \
	fi

test-soak: ## Run 24-hour soak test
	@if command -v k6 > /dev/null; then \
		echo "Running 24-hour soak test..."; \
		echo "This will take approximately 24 hours to complete."; \
		k6 run backend/tests/load/scenarios/soak.js; \
	else \
		echo "Error: k6 is not installed"; \
		echo "Install it with: brew install k6 (macOS) or visit https://k6.io/docs/getting-started/installation/"; \
		exit 1; \
	fi

test-soak-short: ## Run 1-hour soak test (for testing)
	@if command -v k6 > /dev/null; then \
		echo "Running 1-hour soak test..."; \
		k6 run -e DURATION_HOURS=1 backend/tests/load/scenarios/soak.js; \
	else \
		echo "Error: k6 is not installed"; \
		echo "Install it with: brew install k6 (macOS) or visit https://k6.io/docs/getting-started/installation/"; \
		exit 1; \
	fi

test-security: ## Run all security tests (IDOR, authorization)
	@echo "Running IDOR security tests..."
	cd backend && go test -v ./tests/security/
	@echo "Running authorization middleware tests..."
	cd backend && go test -v ./internal/middleware/ -run "TestCanAccessResource|TestPermissionMatrix|TestUserOwnership"
	@echo "✓ All security tests passed"

test-idor: ## Run IDOR vulnerability tests only
	@echo "Running IDOR security tests..."
	cd backend && go test -v ./tests/security/ -run TestIDOR
	@echo "✓ IDOR tests complete"

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf backend/bin
	rm -rf frontend/dist
	@echo "✓ Cleanup complete"

docker-up: ## Start Docker services (PostgreSQL + Redis)
	@echo "Starting Docker services..."
	docker compose -f docker-compose.prod.yml up -d
	@echo "✓ Docker services started"

docker-build: ## Start Docker services (PostgreSQL + Redis)
	@echo "Starting Docker build..."
	docker compose -f docker-compose.prod.yml up -d --build --remove-orphans
	@echo "✓ Docker build complete, and services started"

docker-down: ## Stop Docker services
	@echo "Stopping Docker services..."
	docker compose -f docker-compose.prod.yml down
	@echo "✓ Docker services stopped"

docker-logs: ## View Docker service logs
	@echo "Tailing Docker service logs..."
	docker compose -f docker-compose.prod.yml logs -f --tail 500
	@echo "✓ Docker logs ended"

docker-dev-up: ## Start Docker services for development (PostgreSQL + Redis)
	@echo "Ensuring Docker network $(DEV_NETWORK) exists with correct labels..."
	@if docker network inspect $(DEV_NETWORK) >/dev/null 2>&1; then \
		PROJECT_LABEL=$$(docker network inspect -f '{{ index .Labels "com.docker.compose.project" }}' $(DEV_NETWORK)); \
		NETWORK_LABEL=$$(docker network inspect -f '{{ index .Labels "com.docker.compose.network" }}' $(DEV_NETWORK)); \
		if [ "$$PROJECT_LABEL" != "$(PROJECT_NAME)" ] || [ "$$NETWORK_LABEL" != "default" ]; then \
			echo "Found stale network $(DEV_NETWORK) with mismatched labels ($$PROJECT_LABEL/$$NETWORK_LABEL); recreating..."; \
			docker network rm $(DEV_NETWORK); \
			docker network create --label com.docker.compose.project=$(PROJECT_NAME) --label com.docker.compose.network=default $(DEV_NETWORK); \
		fi; \
	else \
		docker network create --label com.docker.compose.project=$(PROJECT_NAME) --label com.docker.compose.network=default $(DEV_NETWORK); \
	fi
	@echo "Starting Docker services..."
	docker compose -p $(PROJECT_NAME) -f docker-compose.yml up -d
	@echo "✓ Docker services started"

docker-dev-build: ## Build & start Docker services for development (PostgreSQL + Redis)
	@echo "Cleaning up any stale dev containers..."
	@docker compose -p $(PROJECT_NAME) -f docker-compose.yml down --remove-orphans >/dev/null 2>&1 || true
	@echo "Ensuring Docker network $(DEV_NETWORK) exists with correct labels..."
	@if docker network inspect $(DEV_NETWORK) >/dev/null 2>&1; then \
		PROJECT_LABEL=$$(docker network inspect -f '{{ index .Labels "com.docker.compose.project" }}' $(DEV_NETWORK)); \
		NETWORK_LABEL=$$(docker network inspect -f '{{ index .Labels "com.docker.compose.network" }}' $(DEV_NETWORK)); \
		if [ "$$PROJECT_LABEL" != "$(PROJECT_NAME)" ] || [ "$$NETWORK_LABEL" != "default" ]; then \
			echo "Found stale network $(DEV_NETWORK) with mismatched labels ($$PROJECT_LABEL/$$NETWORK_LABEL); recreating..."; \
			docker network rm $(DEV_NETWORK); \
			docker network create --label com.docker.compose.project=$(PROJECT_NAME) --label com.docker.compose.network=default $(DEV_NETWORK); \
		fi; \
	else \
		docker network create --label com.docker.compose.project=$(PROJECT_NAME) --label com.docker.compose.network=default $(DEV_NETWORK); \
	fi
	@echo "Starting Docker build..."
	docker compose -p $(PROJECT_NAME) -f docker-compose.yml up -d --build --remove-orphans
	@echo "✓ Docker build complete, and services started"

docker-dev-down: ## Stop Docker services for development
	@echo "Stopping Docker services..."
	docker compose -f docker-compose.yml down
	@echo "✓ Docker services stopped"

docker-dev-logs: ## View Docker service logs for development
	@echo "Tailing Docker service logs..."
	docker compose -f docker-compose.yml logs -f --tail 500
	@echo "✓ Docker logs ended"

docker-logs-backend: ## Stream backend container logs
	docker logs -f clpr-backend

docker-logs-frontend: ## Stream frontend container logs
	docker logs -f clpr-frontend

docker-logs-postgres: ## Stream postgres container logs
	docker logs -f clpr-postgres

docker-logs-redis: ## Stream redis container logs
	docker logs -f clpr-redis

docker-logs-vault: ## Stream vault-agent container logs
	docker logs -f clpr-vault-agent

backend-dev: ## Run backend in development mode
	@echo "Waiting for PostgreSQL on localhost:5436..."
	@bash -c 'until pg_isready -h localhost -p 5436 -U clpr -d clpr_db >/dev/null 2>&1; do sleep 1; done'
	@echo "PostgreSQL is ready. Starting backend..."
	cd backend && go run ./cmd/api

frontend-dev: ## Run frontend in development mode
	@echo "Starting frontend..."
	cd frontend && npm run dev

backend-build: ## Build backend binary
	@echo "Building backend..."
	cd backend && go build -o bin/api ./cmd/api
	@echo "✓ Backend built"

frontend-build: ## Build frontend for production
	@echo "Building frontend..."
	cd frontend && npm run build
	@echo "✓ Frontend built"

mobile-dev: ## Run mobile app in development mode
	@echo "Starting mobile app..."
	cd mobile && npm start

mobile-ios: ## Run mobile app on iOS simulator
	@echo "Starting mobile app on iOS..."
	cd mobile && npm run ios

mobile-android: ## Run mobile app on Android emulator
	@echo "Starting mobile app on Android..."
	cd mobile && npm run android

mobile-build-ios: ## Build mobile app for iOS
	@echo "Building mobile app for iOS..."
	cd mobile && npm run ios -- --configuration Release
	@echo "✓ Mobile iOS built"

mobile-build-android: ## Build mobile app for Android
	@echo "Building mobile app for Android..."
	cd mobile && npm run android -- --variant=release
	@echo "✓ Mobile Android built"

lint: ## Run linters
	@echo "Linting backend..."
	cd backend && go fmt ./...
	@echo "Linting frontend..."
	cd frontend && npm run lint
	@echo "✓ Linting complete (mobile linting skipped - requires expo CLI fix)"

# Database Migration Commands
DB_URL := "postgresql://clpr:clpr_password@localhost:5436/clpr_db?sslmode=disable"
MIGRATIONS_PATH := backend/migrations

migrate-up: ## Run database migrations up
	@echo "Running database migrations..."
	@if command -v migrate > /dev/null; then \
		migrate -path $(MIGRATIONS_PATH) -database $(DB_URL) up && \
		echo "✓ Migrations completed"; \
	else \
		echo "Error: golang-migrate is not installed"; \
		echo "Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi

migrate-down: ## Rollback last database migration
	@echo "Rolling back database migration..."
	@if command -v migrate > /dev/null; then \
		migrate -path $(MIGRATIONS_PATH) -database $(DB_URL) down 1; \
		echo "✓ Rollback completed"; \
	else \
		echo "Error: golang-migrate is not installed"; \
		exit 1; \
	fi

migrate-down-all: ## Rollback all database migrations
	@echo "Rolling back all database migrations..."
	@if command -v migrate > /dev/null; then \
		migrate -path $(MIGRATIONS_PATH) -database $(DB_URL) down; \
		echo "✓ All migrations rolled back"; \
	else \
		echo "Error: golang-migrate is not installed"; \
		exit 1; \
	fi

migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@if command -v migrate > /dev/null; then \
		migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(NAME); \
		echo "✓ Migration files created"; \
	else \
		echo "Error: golang-migrate is not installed"; \
		exit 1; \
	fi

migrate-status: ## Check current migration version
	@echo "Checking migration status..."
	@if command -v migrate > /dev/null; then \
		migrate -path $(MIGRATIONS_PATH) -database $(DB_URL) version; \
	else \
		echo "Error: golang-migrate is not installed"; \
		exit 1; \
	fi

migrate-seed: ## Seed database with sample data
	@echo "Seeding database..."
	@PGPASSWORD=clpr_password psql -h localhost -p 5436 -U clpr -d clpr_db -f $(MIGRATIONS_PATH)/seed.sql
	@echo "✓ Database seeded"

migrate-seed-load-test: ## Seed database with load test data (includes sample data)
	@echo "Seeding database with load test data..."
	@PGPASSWORD=clpr_password psql -h localhost -p 5436 -U clpr -d clpr_db -f $(MIGRATIONS_PATH)/seed.sql
	@PGPASSWORD=clpr_password psql -h localhost -p 5436 -U clpr -d clpr_db -f $(MIGRATIONS_PATH)/seed_load_test.sql
	@echo "✓ Load test data seeded"

migrate-seed-moderation-perf-test: ## Seed database with moderation performance test data (10K+ bans, 50K+ audit logs)
	@echo "Seeding database with moderation performance test data..."
	@echo "This will create:"
	@echo "  - 12,000+ ban records"
	@echo "  - 55,000+ audit log entries"
	@echo "  - 5,000+ moderation queue items"
	@echo "  - 100+ community moderators"
	@PGPASSWORD=clpr_password psql -h localhost -p 5436 -U clpr -d clpr_db -f $(MIGRATIONS_PATH)/seed_moderation_perf_test.sql
	@echo "✓ Moderation performance test data seeded successfully"

site-freshness-seed: ## Ensure default public smart playlists exist for fresh site content
	@echo "Ensuring default site freshness playlist rules exist..."
	@cd backend && go run ./cmd/seed-site-freshness
	@echo "✓ Site freshness rules ensured"

site-freshness-generate: ## Ensure default smart playlists exist and generate a fresh batch immediately
	@echo "Ensuring site freshness rules exist and generating playlists now..."
	@cd backend && go run ./cmd/seed-site-freshness -generate-now
	@echo "✓ Site freshness rules ensured and generated"

# Search Evaluation
evaluate-search: ## Run search quality evaluation
	@echo "Running search evaluation..."
	@cd backend && go build -o bin/evaluate-search ./cmd/evaluate-search
	@cd backend && ./bin/evaluate-search -verbose
	@echo "✓ Search evaluation complete"

evaluate-search-json: ## Run search evaluation and output JSON
	@echo "Running search evaluation..."
	@cd backend && go build -o bin/evaluate-search ./cmd/evaluate-search
	@cd backend && ./bin/evaluate-search -output evaluation-results.json
	@echo "✓ Results saved to backend/evaluation-results.json"

# Recommendation Evaluation
evaluate-recommendations: ## Run recommendation quality evaluation
	@echo "Running recommendation evaluation..."
	@cd backend && go build -o bin/evaluate-recommendations ./cmd/evaluate-recommendations
	@cd backend && ./bin/evaluate-recommendations -verbose
	@echo "✓ Recommendation evaluation complete"

evaluate-recommendations-json: ## Run recommendation evaluation and output JSON
	@echo "Running recommendation evaluation..."
	@cd backend && go build -o bin/evaluate-recommendations ./cmd/evaluate-recommendations
	@cd backend && ./bin/evaluate-recommendations -output recommendation-evaluation-results.json
	@echo "✓ Results saved to backend/recommendation-evaluation-results.json"

grid-search-recommendations: ## Run parameter grid search for recommendation tuning
	@echo "Running recommendation parameter grid search..."
	@cd backend && go build -o bin/grid-search-recommendations ./cmd/grid-search-recommendations
	@cd backend && ./bin/grid-search-recommendations -quick -verbose
	@echo "✓ Grid search complete"

grid-search-recommendations-full: ## Run full parameter grid search (slower but more thorough)
	@echo "Running full recommendation parameter grid search..."
	@cd backend && go build -o bin/grid-search-recommendations ./cmd/grid-search-recommendations
	@cd backend && ./bin/grid-search-recommendations -output grid-search-results.json
	@echo "✓ Full grid search complete - results saved to backend/grid-search-results.json"

# Kubernetes targets
k8s-provision: ## Provision a new Kubernetes cluster (set CLOUD_PROVIDER, CLUSTER_NAME, REGION, etc.)
	@echo "Provisioning Kubernetes cluster..."
	@./infrastructure/k8s/bootstrap/provision-cluster.sh
	@echo "✓ Cluster provisioned"

k8s-setup: ## Set up cluster with required operators and configuration
	@echo "Setting up cluster components..."
	@./infrastructure/k8s/bootstrap/setup-cluster.sh
	@echo "✓ Cluster setup complete"

k8s-verify: ## Verify cluster health and configuration
	@echo "Verifying cluster health..."
	@./infrastructure/k8s/bootstrap/verify-cluster.sh

k8s-deploy-prod: ## Deploy applications to production namespace
	@echo "Deploying to production..."
	@kubectl apply -k infrastructure/k8s/overlays/production/
	@kubectl rollout status deployment/clpr-backend -n clpr-production
	@echo "✓ Production deployment complete"

k8s-deploy-staging: ## Deploy applications to staging namespace
	@echo "Deploying to staging..."
	@kubectl apply -k infrastructure/k8s/overlays/staging/
	@kubectl rollout status deployment/clpr-backend -n clpr-staging
	@echo "✓ Staging deployment complete"

k8s-logs-prod: ## View backend logs in production
	@kubectl logs -f -l app=clpr-backend -n clpr-production

k8s-logs-staging: ## View backend logs in staging
	@kubectl logs -f -l app=clpr-backend -n clpr-staging

k8s-status-prod: ## Show status of production deployment
	@echo "=== Production Status ==="
	@kubectl get pods -n clpr-production
	@kubectl get ingress -n clpr-production
	@kubectl get certificate -n clpr-production

k8s-status-staging: ## Show status of staging deployment
	@echo "=== Staging Status ==="
	@kubectl get pods -n clpr-staging
	@kubectl get ingress -n clpr-staging
	@kubectl get certificate -n clpr-staging

# OpenAPI Documentation
openapi-validate: ## Validate OpenAPI specification
	@echo "Validating OpenAPI specification..."
	npm run openapi:validate

openapi-serve: ## Start Swagger UI server on http://localhost:8081
	@echo "Starting Swagger UI server..."
	@echo "Open http://localhost:8081 in your browser"
	npm run openapi:serve

openapi-build: ## Build static API documentation
	@echo "Building static API documentation..."
	npm run openapi:build
	@echo "✓ Documentation built at docs/openapi/api-docs.html"

openapi-preview: ## Preview OpenAPI docs with live reload
	@echo "Starting Redocly preview server..."
	npm run openapi:preview

openapi-stats: ## Show OpenAPI spec statistics
	@echo "Analyzing OpenAPI specification..."
	npm run openapi:stats

# =============================================================================
# VPS Deployment Targets
# =============================================================================

deploy-vps: ## Deploy to VPS (production) with Vault + Caddy
	@echo "Starting VPS deployment..."
	./scripts/deploy-vps.sh

deploy-vps-status: ## Check VPS deployment status
	@echo "Checking VPS deployment status..."
	@docker compose -p $(PROJECT_NAME) -f docker-compose.vps.yml ps
	@echo ""
	@echo "Network connectivity (web):"
	@docker network inspect web --format '{{range .Containers}}{{.Name}} {{end}}' || echo "Network 'web' not found"

deploy-vps-logs: ## View VPS deployment logs
	@docker compose -p $(PROJECT_NAME) -f docker-compose.vps.yml logs -f

deploy-vps-down: ## Stop VPS deployment
	@echo "Stopping VPS deployment..."
	@docker compose -p $(PROJECT_NAME) -f docker-compose.vps.yml down
