#!/usr/bin/make -f

VERSION = $(shell git rev-parse HEAD)

.PHONY: help
help:
	@echo
	@echo "Commands"
	@echo "========"
	@echo
	@sed -n '/^[a-zA-Z0-9_-]*:/s/:.*//p' < Makefile | grep -v -E 'default|help.*' | sort


.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags "-extldflags '-static'" -o build/ ./cmd/...


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
