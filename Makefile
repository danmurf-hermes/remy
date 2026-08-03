.PHONY: build dev run setup test lint lint-go lint-frontend fmt clean pre-commit ci ci-check release release-dry-run

BINARY_NAME=remy
BUILD_DIR=build
VERSION=$(shell git describe --tags 2>/dev/null || echo "dev")
LDFLAGS=-ldflags="-X main.Version=$(VERSION) -s -w"

# First-time setup: install Wails CLI and frontend deps
setup:
	@echo "Installing Wails CLI..."
	@go install github.com/wailsapp/wails/v2/cmd/wails@latest
	@echo "Installing frontend dependencies..."
	@cd frontend && npm ci
	@echo "Done. Run 'make run' to start the app."

# One-command run: ensures deps are installed, then launches the GUI
run: frontend-deps
	@PATH="$$(go env GOPATH)/bin:$$PATH" wails dev

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) .

dev:
	wails dev

test:
	go test ./internal/... -cover
	cd frontend && npm test

lint: lint-go lint-frontend

lint-go:
	golangci-lint run ./...

lint-frontend:
	cd frontend && npx eslint --ext .js,.svelte src/
	cd frontend && npx prettier --check src/

fmt:
	gofmt -s -w .
	cd frontend && npx prettier --write src/

pre-commit: fmt lint test

# Mirrors .github/workflows/ci.yml exactly (ci job, in order):
# frontend deps -> frontend build -> Go lint -> Go test -> frontend lint -> frontend format check -> frontend test -> Go build
# Run this BEFORE every push. Do not push until it passes.
ci: frontend-deps frontend-build lint-go go-test lint-frontend frontend-test go-build

ci-check:
	@echo "=== CI pipeline (mirrors .github/workflows/ci.yml) ==="
	@echo "1. Frontend dependencies (npm ci)"
	@echo "2. Frontend build (npm run build)"
	@echo "3. Go lint (golangci-lint)"
	@echo "4. Go test (go test ./internal/... -cover)"
	@echo "5. Frontend lint (ESLint)"
	@echo "6. Frontend format check (Prettier)"
	@echo "7. Frontend test (Vitest)"
	@echo "8. Go build"
	@echo "Run 'make ci' to execute the full pipeline."

go-test:
	go test ./internal/... -cover -coverprofile=coverage.out

go-build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) .

frontend-test:
	cd frontend && npm test

frontend-deps:
	cd frontend && npm ci

frontend-build:
	cd frontend && npm run build

clean:
	rm -rf $(BUILD_DIR)
	rm -rf frontend/node_modules
	rm -rf frontend/dist

# Cross-platform release builds
release: frontend-deps frontend-build
	@mkdir -p $(BUILD_DIR)/release
	@echo "Building for linux/amd64..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 .
	@echo "Building for linux/arm64..."
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-arm64 .
	@echo "Building for darwin/amd64..."
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 .
	@echo "Building for darwin/arm64..."
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64 .
	@echo "Building for windows/amd64..."
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Release builds complete:"
	@ls -lh $(BUILD_DIR)/release/

release-dry-run:
	@echo "=== Dry run: would build for all platforms ==="
	@echo "Targets: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64"
	@echo "Version: $(VERSION)"
	@echo "LDFLAGS: $(LDFLAGS)"
	@echo "Run 'make release' to actually build."
