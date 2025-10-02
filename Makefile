.PHONY: build clean test lint

VERSION ?= 0.1.0
COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS = -X main.buildVersion=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}

build:  ## Build the binary with version info
	mkdir -p dist
	go build -ldflags "${LDFLAGS}" -o dist/moxli ./cmd/moxli

clean:  ## Clean build artifacts
	rm -rf dist/
	go clean -cache

test:  ## Run all tests
	go test -v ./...

lint:  ## Run golangci-lint
	golangci-lint run ./...

fmt:  ## Format all Go files
	go fmt ./...
	goimports -local "github.com/lelopez-io/moxli" -w .

install:  ## Install moxli to GOPATH/bin
	go install -ldflags "${LDFLAGS}" ./cmd/moxli
