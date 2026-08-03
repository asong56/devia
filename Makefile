.PHONY: build build-min build-all clean size vet test help

build: ## Full binary (CLI + `devia serve`) for the current platform, stripped
	go build -trimpath -ldflags="-s -w" -o devia .

build-min: ## CLI-only binary (-tags noserve), no net/http linked, smallest size
	go build -tags noserve -trimpath -ldflags="-s -w" -o devia-cli .

build-all: ## Cross-compile both variants for linux/windows/macos, amd64+arm64
	bash build.sh

vet: ## Run go vet (static analysis) across the whole module
	go vet ./...

test: ## Run the test suite (core package)
	go test ./... -v

size: build build-min ## Build both variants and print their sizes
	@echo "full (CLI+API): $$(du -h devia | cut -f1)"
	@echo "cli-only:       $$(du -h devia-cli | cut -f1)"

clean: ## Remove all build output
	rm -rf devia devia-cli devia.exe devia-cli.exe build/

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-14s\033[0m %s\n", $$1, $$2}'
