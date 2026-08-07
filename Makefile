.PHONY: build build-min build-all test test-race clean size vet help

test: ## Run the test suite (internal/core, internal/cli, cmd/devia black-box)
	go test -count=1 ./...

test-race: ## Same as test, with the race detector enabled
	go test -race -count=1 ./...

build: test ## Full binary (CLI + `devia serve`) for the current platform, stripped
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o devia ./cmd/devia

build-min: test ## CLI-only binary (-tags noserve), no net/http linked, smallest size
	CGO_ENABLED=0 go build -tags noserve -trimpath -ldflags="-s -w" -o devia-cli ./cmd/devia

build-all: test ## Cross-compile both variants for linux/windows/macos, amd64+arm64
	bash build.sh

vet: ## Run go vet (static analysis) across the whole module
	go vet ./...

size: build build-min ## Build both variants and print their sizes
	@echo "full (CLI+API): $$(du -h devia | cut -f1)"
	@echo "cli-only:       $$(du -h devia-cli | cut -f1)"

clean: ## Remove all build output
	rm -rf devia devia-cli devia.exe devia-cli.exe build/ dist/

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-14s\033[0m %s\n", $$1, $$2}'
