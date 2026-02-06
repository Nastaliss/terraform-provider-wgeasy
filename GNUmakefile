NAME                := terraform-provider-wgeasy
VERSION             := $(shell git describe --tags --abbrev=1 2>/dev/null || echo "dev")
FILES               := $(shell git ls-files '*.go' 'internal/**/*.go')
DEV_REPOSITORY_PATH := registry.terraform.io/Nastaliss/wgeasy
DEV_VERSION         := 0.1.0
OS_ARCH             := linux_amd64

.DEFAULT_GOAL := help

.PHONY: setup
setup: ## Install required libraries/tools for build tasks
	@command -v goimports 2>&1 >/dev/null    || go install golang.org/x/tools/cmd/goimports@latest
	@command -v gosec 2>&1 >/dev/null        || go install github.com/securego/gosec/v2/cmd/gosec@latest
	@command -v goveralls 2>&1 >/dev/null    || go install github.com/mattn/goveralls@latest
	@command -v ineffassign 2>&1 >/dev/null  || go install github.com/gordonklaus/ineffassign@latest
	@command -v misspell 2>&1 >/dev/null     || go install github.com/client9/misspell/cmd/misspell@latest
	@command -v revive 2>&1 >/dev/null       || go install github.com/mgechev/revive@latest
	@command -v tfplugindocs 2>&1 >/dev/null || go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest

.PHONY: fmt
fmt: setup ## Format source code
	goimports -w $(FILES)

.PHONY: lint
lint: revive vet goimports ineffassign misspell gosec ## Run all lint related tests against the codebase

.PHONY: revive
revive: setup ## Test code syntax with revive
	revive -config .revive.toml $(FILES)

.PHONY: vet
vet: ## Test code syntax with go vet
	go vet ./...

.PHONY: goimports
goimports: setup ## Test code syntax with goimports
	goimports -d $(FILES) > goimports.out
	@if [ -s goimports.out ]; then cat goimports.out; rm goimports.out; exit 1; else rm goimports.out; fi

.PHONY: ineffassign
ineffassign: setup ## Test code syntax for ineffassign
	ineffassign ./...

.PHONY: misspell
misspell: setup ## Test code with misspell
	misspell -error $(FILES)

.PHONY: gosec
gosec: setup ## Test code for security vulnerabilities
	gosec ./...

.PHONY: test
test: ## Run the tests against the codebase
	go test -v -race ./...

.PHONY: testacc
testacc: ## Run acceptance tests
	TF_ACC=1 go test -v -race ./... -timeout 120m

.PHONY: docs
docs: setup ## Generate documentation
	tfplugindocs

.PHONY: build
build: ## Build the binary
	go build -o $(NAME)

.PHONY: install
install: build ## Build and install the provider locally
	mkdir -p ~/.terraform.d/plugins/$(DEV_REPOSITORY_PATH)/$(DEV_VERSION)/$(OS_ARCH)
	mv $(NAME) ~/.terraform.d/plugins/$(DEV_REPOSITORY_PATH)/$(DEV_VERSION)/$(OS_ARCH)/

.PHONY: release-snapshot
release-snapshot: ## Build release binaries (snapshot, no publish)
	goreleaser release --snapshot --skip=publish --skip=sign --clean

.PHONY: release
release: ## Build & release the binaries
	goreleaser release --clean

.PHONY: clean
clean: ## Remove binary if it exists
	rm -f $(NAME)

.PHONY: coverage
coverage: ## Generate coverage report
	rm -rf *.out
	go test -v ./... -coverpkg=./... -coverprofile=coverage.out

.PHONY: all
all: lint test build coverage ## Test, build and generate coverage

.PHONY: help
help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
