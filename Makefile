.PHONY: build test clean vendor

# Binary name
BINARY_NAME=pull-metrics

# Build the application
build:
	go build -buildvcs=false -o $(BINARY_NAME) .

# Run tests
test:
	go test -v ./...

# Run a specific test
test-single:
	@if [ -z "$(TEST)" ]; then \
		echo "Usage: make test-single TEST=TestFunctionName"; \
		exit 1; \
	fi
	go test -v -run $(TEST) ./...

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)

# Vendor dependencies
vendor:
	go mod tidy
	go mod vendor

# Install dependencies
deps:
	go mod download

# Format code
fmt:
	go fmt ./...

# Lint code (if golint is available)
lint:
	@which golint > /dev/null || (echo "golint not found, install with: go install golang.org/x/lint/golint@latest"; exit 1)
	golint ./...

# Run all checks
check: fmt lint test

# Help
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  test       - Run all tests"
	@echo "  test-single TEST=name - Run a specific test"
	@echo "  clean      - Clean build artifacts"
	@echo "  vendor     - Vendor dependencies"
	@echo "  deps       - Download dependencies"
	@echo "  fmt        - Format code"
	@echo "  lint       - Lint code"
	@echo "  check      - Run fmt, lint, and test"
	@echo "  help       - Show this help"