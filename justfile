# Run the test suite
test:
    go test ./...

# Format the code
fmt:
    golangci-lint fmt

# Lint the code, applying fixes where possible
lint:
    golangci-lint run --fix

# Report suspicious constructs
vet:
    go vet ./...

# Apply fixes for outdated APIs
fix:
    go fix ./...
