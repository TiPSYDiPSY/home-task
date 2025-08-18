GO_VERSION = $(shell sed -En 's/^go (.*)$$/\1/p' go.mod)
VERSION ?= $(shell cat .version 2> /dev/null)
DATE    ?= $(shell date +%FT%T%z)

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

build:
	go build \
		-tags release \
		-ldflags '-X home-task/cmd.Version=$(VERSION) -X home-task/cmd.BuildDate=$(DATE)' \
		-o bin/home-task cmd/home-task/main.go

run:
	bin/home-task