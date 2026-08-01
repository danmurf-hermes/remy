.PHONY: build dev test lint clean

BINARY_NAME=remy
BUILD_DIR=build
VERSION=$(shell git describe --tags 2>/dev/null || echo "dev")

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/remy

dev:
	wails dev

test:
	go test ./internal/... -cover
	cd frontend && npm test

lint:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
	rm -rf frontend/node_modules
	rm -rf frontend/dist
