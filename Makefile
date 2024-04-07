.PHONY: vet fmt test

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

vet: ## Run go vet on all packages
	go vet ./...

fmt: ## Run go fmt on all packages
	go fmt ./...

test: ## Run go test on all packages
	go test -v ./...
