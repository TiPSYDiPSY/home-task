GO_VERSION = $(shell sed -En 's/^go (.*)$$/\1/p' go.mod)

format: ## formats the source code
	@go fmt ./...

install-tools:
	@echo Installing tools
	go install tool
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.6

lint: ## lints the source code
	golangci-lint run --go $(GO_VERSION) --timeout 2m0s -v

test: ## execute unit tests on local env
	go test ./...

generate: ## generate source code (mocks, enums)
	mockery