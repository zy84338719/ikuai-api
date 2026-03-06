.PHONY: test test-coverage test-integration lint fmt vet build example clean

# Run unit tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run integration tests (requires real iKuai router)
test-integration:
	@echo "Running integration tests..."
	go test -tags=integration -v ./...

# Run integration tests with custom config
test-integration-custom:
	@echo "Running integration tests with custom config..."
	IKUAI_TEST_ADDR=$(ADDR) IKUAI_TEST_USERNAME=$(USER) IKUAI_TEST_PASSWORD=$(PASS) \
		go test -tags=integration -v ./...

# Run linter
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Build the SDK
build:
	go build ./...

# Run example (requires real iKuai router)
example:
	cd example && go run main.go

# Clean generated files
clean:
	rm -f coverage.out coverage.html

# Run all checks
check: fmt vet test
	@echo "All checks passed!"

# Install dependencies
deps:
	go mod download
	go mod tidy

# Generate documentation
doc:
	go doc -all ./...

# Show help
help:
	@echo "iKuai SDK Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make test              Run unit tests"
	@echo "  make test-coverage     Run tests with coverage report"
	@echo "  make test-integration  Run integration tests (requires router)"
	@echo "  make lint              Run linter"
	@echo "  make fmt               Format code"
	@echo "  make vet               Run go vet"
	@echo "  make build             Build the SDK"
	@echo "  make example           Run the example"
	@echo "  make check             Run all checks (fmt, vet, test)"
	@echo "  make deps              Install dependencies"
	@echo "  make doc               Generate documentation"
	@echo "  make help              Show this help"
