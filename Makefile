#!/usr/bin/make -f

export VERSION := $(shell echo $(shell git describe --tags --always --match "v*") | sed 's/^v//')

.PHONY: help
help:
	@echo
	@echo "Commands"
	@echo "========"
	@echo
	@sed -n '/^[a-zA-Z0-9_-]*:/s/:.*//p' < Makefile | grep -v -E 'default|help.*' | sort


.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags "-extldflags '-static' -X main.Version=$(VERSION)"  -o build/ ./cmd/...

.PHONY: build-linux-amd64
build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-extldflags '-static' -X main.Version=$(VERSION)"  -o build/opull-$(VERSION)-linux-amd64 ./cmd/...

.PHONY: test
test: test-unit

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test-unit
test-unit: fmt vet
	@VERSION=$(VERSION) go test -mod=readonly ./...

.PHONY: test-race
test-race: fmt vet
	@VERSION=$(VERSION) go test -race -mod=readonly ./...

.PHONY: test-all
test-all: test-race
