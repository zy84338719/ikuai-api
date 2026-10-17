.PHONY: test test-coverage lint fmt vet build example generate check help

# Run unit tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter (best-effort; does not fail the build if golangci-lint is missing)
lint:
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Build the SDK
build:
	go build ./...

# Regenerate the service layer from v4_catalog.go
generate:
	go run ./codegen

# Run the example (uses IKUAI_DRY_RUN=1 to avoid router traffic)
example:
	cd example && IKUAI_DRY_RUN=1 go run main.go

# Clean generated files
clean:
	rm -f coverage.out coverage.html

# Run all standard checks
check: fmt vet test zero-dep
	@echo "All checks passed!"

# Show help
help:
	@echo "ikuai-api Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make test          Run unit tests"
	@echo "  make test-coverage Run tests with coverage"
	@echo "  make generate      Regenerate service/ from v4_catalog.go"
	@echo "  make lint          Run linter (if installed)"
	@echo "  make fmt           Format code"
	@echo "  make vet           Run go vet"
	@echo "  make build         Build the SDK"
	@echo "  make example       Run example in dry-run mode"
	@echo "  make check         Run fmt, vet and test"

# Dependency gate: this SDK may depend on the oekit modules and nothing
# else. A new third-party require is a design decision, not an accident, so
# it has to fail loudly rather than slip into a tidy.
zero-dep:
	@bad=$$(go list -m all | grep -v '^github.com/zy84338719/ikuai-api' \
		| grep -v '^github.com/oekit/' \
		| grep -v '^$$' || true); \
	if [ -n "$$bad" ]; then \
		echo "dependency gate failed — only github.com/oekit/* is allowed:"; \
		echo "$$bad"; exit 1; \
	fi; \
	echo "dependency gate ok: no third-party requires; oekit modules:"; \
		go list -m all | grep '^github.com/oekit/' | sed 's/^/  /'

.PHONY: zero-dep
