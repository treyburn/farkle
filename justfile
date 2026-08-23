bin_dir := justfile_directory() / "bin"
data_dir := justfile_directory() / "data"
pkg := "."

# Prints all available recipes.
help:
    @just --list

# Build for Linux (amd64).
build: build-linux

# Build for Linux (amd64).
build-linux:
    mkdir -p {{bin_dir}}
    GOOS=linux GOARCH=amd64 go build -o {{bin_dir}}/farkle-linux-amd64 {{pkg}}

# Build for macOS (arm64).
build-mac:
    mkdir -p {{bin_dir}}
    GOOS=darwin GOARCH=arm64 go build -o {{bin_dir}}/farkle-darwin-arm64 {{pkg}}

# Build for Windows (amd64).
build-windows:
    mkdir -p {{bin_dir}}
    GOOS=windows GOARCH=amd64 go build -o {{bin_dir}}/farkle-windows-amd64.exe {{pkg}}

# Build every supported platform.
build-all: build-linux build-mac build-windows

# Remove the build output.
clean:
    rm -rf {{bin_dir}}

# Rebuild data/dice.json from the xml files already in ./data.
data-regenerate:
    nu {{data_dir}}/process.nu

# Extract the source xml from a KCD2 install, then rebuild data/dice.json. Must provide the path to your game data.
data-extract game_dir:
    nu {{data_dir}}/extract.nu {{game_dir}}

# Run the test suite with the race detector and per-package coverage.
test:
    go test -race -coverprofile=coverage.txt -covermode=atomic ./...

# Open the coverage profile from `just test` in a browser.
cover: test
    go tool cover -html=coverage.txt

# Format the code.
fmt:
    golangci-lint fmt

# Lint the code, applying fixes where possible.
lint:
    golangci-lint run --fix

# Identify insecure code.
vet:
    go vet ./...

# Apply fixes for outdated APIs.
fix:
    go fix ./...

# A quick and convenient wrapper to run a bunch of common commands before CI.
ci: fix fmt vet lint test
