GO ?= go
GOLANGCI_LINT ?= golangci-lint
VERSION ?= dev
export GOWORK = off
export GOTOOLCHAIN = auto

.PHONY: build test vet lint check
build:
	$(GO) build -trimpath -ldflags "-X main.version=$(VERSION)" -o bin/az-vote-delegate ./cmd/az-vote-delegate

test:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

lint:
	$(GOLANGCI_LINT) run ./...

check: test vet lint
