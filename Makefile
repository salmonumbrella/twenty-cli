.PHONY: build install clean test lint fmt help

BINARY_NAME := twenty
MODULE := github.com/salmonumbrella/twenty-cli
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

## help: Show this help
help:
	@grep -E '^##' $(MAKEFILE_LIST) | sed -e 's/^##//'

## build: Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/$(BINARY_NAME)

## install: Install the binary to GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/$(BINARY_NAME)

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

## test: Run tests (uses file backend to avoid keychain popups)
test:
	TWENTY_KEYRING_BACKEND=file TWENTY_KEYRING_PASSWORD=test go test -v -race -cover ./...

## lint: Run linters
lint:
	golangci-lint run

## fmt: Format code
fmt:
	go fmt ./...
	goimports -w .

## tidy: Tidy dependencies
tidy:
	go mod tidy

## release-snapshot: Build snapshot release (no publish)
release-snapshot:
	goreleaser release --snapshot --clean

## release: Build and publish release
release:
	goreleaser release --clean
