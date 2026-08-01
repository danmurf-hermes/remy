.PHONY: build dev test lint lint-go lint-frontend fmt clean pre-commit

BINARY_NAME=remy
BUILD_DIR=build
VERSION=$(shell git describe --tags 2>/dev/null || echo "dev")

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

clean:
	rm -rf $(BUILD_DIR)
	rm -rf frontend/node_modules
	rm -rf frontend/dist
