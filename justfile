# Run the test suite with the race detector and per-package coverage
test:
    go test -race -coverprofile=coverage.txt -covermode=atomic ./...

# Open the coverage profile from `just test` in a browser
cover: test
    go tool cover -html=coverage.txt

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
