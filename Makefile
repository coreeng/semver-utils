.PHONY: vet fmt test test-go test-scripts lint build

# Default goreleaser command
GORELEASER_CMD ?= goreleaser

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

fmt: ## Format using golangci-lint
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@golangci-lint fmt || exit 1

lint: ## Lint using golangci-lint
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@golangci-lint run || exit 1

test-go: ## Run Go unit tests on all packages
	go test -v ./...

test-scripts: ## Run integration tests from the tests directory
	@if [ ! -d "tests" ]; then \
		echo "Error: tests directory not found"; \
		exit 1; \
	fi
	@cd tests && ./test_semver.sh && ./test_semver-git.sh

test: test-go test-scripts ## Run all tests (both Go unit tests and integration tests)

build: ## Build the project using goreleaser in snapshot mode
	$(GORELEASER_CMD) build --snapshot --clean --single-target

release: ## Release the project using goreleaser
	$(GORELEASER_CMD) release --clean
