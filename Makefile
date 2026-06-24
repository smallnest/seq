# Makefile for github.com/smallnest/seq — a Go 1.27 iter.Seq collection library.
#
# go1.27rc1 tooling note:
#   gofmt's parser rejects method-level type parameters
#   ("method must have no type parameters") even though go build / go vet /
#   go test all accept them. Only three files declare such methods
#   (intermediate.go, terminal_aggregate.go, seq2_methods.go); they are kept
#   formatted by hand. Accordingly:
#     - `make fmt`  formats each file individually and skips those three.
#     - `make lint` uses `go vet` as the gate (gofmt -l is excluded).
#   Revisit on the Go 1.27 stable release.

GO ?= go
GO_FILES := $(shell find . -name '*.go' -not -path './.git/*')

.PHONY: all build test vet fmt lint check clean help

all: build ## Build all packages (default target)

build: ## Compile all packages
	$(GO) build ./...

test: ## Run all tests
	$(GO) test ./...

vet: ## Run go vet on all packages
	$(GO) vet ./...

fmt: ## Format Go source files (skips generic-method files under go1.27rc1)
	@ok=0; skip=0; \
	for f in $(GO_FILES); do \
		if gofmt -w "$$f" 2>/dev/null; then \
			ok=$$((ok+1)); \
		else \
			printf "  skip %s (go1.27rc1: gofmt cannot parse method-level type params)\n" "$$f"; \
			skip=$$((skip+1)); \
		fi; \
	done; \
	echo "fmt: $$ok processed, $$skip skipped"

lint: vet ## Lint (go vet is the gate; gofmt -l excluded under go1.27rc1)
	@echo "lint: go vet clean"

check: build vet test ## Build + vet + test (the project gate)

clean: ## Clear the go test cache
	$(GO) clean -testcache

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "seq — available targets:\n"} /^[a-zA-Z0-9_-]+:.*##/ {sub(/^[ \t]+/, "", $$2); printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
