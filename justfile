# Gwaihir - WoL Messenger Service
# Go-based microservice for sending Wake-on-LAN packets

# Default recipe to display help information
default:
    @just --list

# Install dependencies (for CI)
ci:
    go mod download
    go mod verify

# Run all checks (format and lint)
check: format-check lint

# Format code
format:
    gofmt -w .
    goimports -w -local github.com/josimar-silva/gwaihir .

# Check if code is formatted
format-check:
    #!/usr/bin/env bash
    set -euo pipefail

    unformatted=$(gofmt -l .)
    if [ -n "$unformatted" ]; then
        echo "The following files are not formatted:"
        echo "$unformatted"
        exit 1
    fi
    echo "All files are properly formatted"

# Run linters
lint:
    golangci-lint run ./...

# Run tests with coverage
test:
    go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
    go tool cover -html=coverage.out -o coverage.html

# Run tests without coverage (faster)
test-fast:
    go test -v -race ./...

# Run integration tests
test-integration:
    go test -v -tags=integration -run TestIntegration ./tests/...

# Build the binary
build:
    go build -o bin/gwaihir -ldflags="-s -w" ./cmd/gwaihir

# Build for release (with version info)
build-release VERSION:
    #!/usr/bin/env bash
    set -euo pipefail

    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    GIT_COMMIT=$(git rev-parse --short HEAD)

    go build \
        -o bin/gwaihir \
        -ldflags="-s -w -X main.Version={{VERSION}} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        ./cmd/gwaihir

# Run the application locally
run:
    go run ./cmd/gwaihir

# Clean build artifacts
clean:
    rm -rf bin/
    rm -f coverage.out coverage.html

# Run all quality checks before committing
pre-commit: format lint test test-integration

# Prepare for a new release
pre-release:
    #!/usr/bin/env bash
    set -euo pipefail

    if [[ -n $(git status --porcelain) ]]; then
        echo "Git working directory is not clean. Please commit or stash your changes."
        exit 1
    fi

    echo "Running checks and tests..."
    just lint
    just test

    current_version=$(grep '^const Version' cmd/gwaihir/version.go | cut -d '"' -f 2)
    echo "Current version is ${current_version}"

    if [[ $current_version != *"-SNAPSHOT"* ]]; then
        echo "Error: Current version is not a SNAPSHOT version."
        exit 1
    fi

    new_version=$(echo ${current_version} | sed 's/-SNAPSHOT//')
    echo "Bumping version to ${new_version}..."

    # Update version in version.go
    sed -i "s/const Version = \".*\"/const Version = \"${new_version}\"/" cmd/gwaihir/version.go

    # Update version in README if it exists
    if [ -f README.md ]; then
        sed -i "s/gwaihir-v[0-9.]*/gwaihir-v${new_version}/g" README.md
    fi

    echo "Committing version bump..."
    git add cmd/gwaihir/version.go README.md
    git commit -m "chore(release): prepare for release v${new_version}"

    echo "Pre-release for version ${new_version} is ready."
    echo "You can now push the changes to trigger the release workflow."

# Update dependencies
deps-update:
    go get -u ./...
    go mod tidy

# Generate mocks (if you use mockgen)
mocks:
    @echo "No mocks to generate yet"

# Run security checks
security:
    gosec -quiet ./...

# Show project statistics
stats:
    @echo "Lines of code:"
    @find . -name '*.go' -not -path "./vendor/*" | xargs wc -l | tail -1
    @echo ""
    @echo "Test coverage:"
    @go test -cover ./... | grep coverage

# Docker build
docker-build TAG="latest":
    docker build -t ghcr.io/josimar-silva/gwaihir:{{TAG}} .

# Docker push
docker-push TAG="latest":
    docker push ghcr.io/josimar-silva/gwaihir:{{TAG}}
