SHELL = /usr/bin/env bash -o nounset -o pipefail -o errexit -c

ifneq ("$(wildcard .env)","")
	include .env
	export
endif

APP_NAME = http
PACKAGES ?= ./...

MAIN_SOURCE = ./cmd/http
MAIN_RUNNER = go run $(MAIN_SOURCE)
ifeq ($(DEBUG), true)
	MAIN_RUNNER = dlv debug $(MAIN_SOURCE) --
endif

.DEFAULT_GOAL := app

## help: Display list of commands
.PHONY: help
help: Makefile
	@sed -n 's|^##||p' $< | column -t -s ':' | sort

## name: Output app name
.PHONY: name
name:
	@printf "$(APP_NAME)"

## version: Output last commit sha
.PHONY: version
version:
	@printf "$(shell git rev-parse --short HEAD)"

## version-date: Output last commit date
.PHONY: version-date
version-date:
	@printf "$(shell git log -n 1 "--date=format:%Y%m%d%H%M" "--pretty=format:%cd")"

## dev: Build app
.PHONY: dev
dev: format style test build

## app: Build whole app
.PHONY: app
app: init dev

## init: Bootstrap your application. e.g. fetch some data files, make some API calls, request user input etc...
.PHONY: init
init:
	@curl --disable --silent --show-error --location "https://raw.githubusercontent.com/ViBiOh/scripts/main/bootstrap" | bash -s -- "-c" "git_hooks" "coverage" "release"
	go install "github.com/golang/mock/mockgen@v1.6.0"
	go install "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
	go install "golang.org/x/tools/cmd/goimports@latest"
	go install "golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@master"
	go install "mvdan.cc/gofumpt@latest"
	$(MAKE) mocks
	go mod tidy

## format: Format code. e.g Prettier (js), format (golang)
.PHONY: format
format:
	find . -name "*.go" -exec goimports -w {} \+
	find . -name "*.go" -exec gofumpt -w {} \+

## style: Check lint, code styling rules. e.g. pylint, phpcs, eslint, style (java) etc ...
.PHONY: style
style:
	fieldalignment -fix -test=false $(PACKAGES)
	golangci-lint run

## mocks: Generate mocks
.PHONY: mocks
mocks:
	find . -name "mocks" -type d -exec rm -r "{}" \+
	go generate -run mockgen $(PACKAGES)
	mockgen -destination pkg/mocks/io.go -package mocks -mock_names ReadCloser=ReadCloser io ReadCloser
	mockgen -destination pkg/mocks/pgx.go -package mocks -mock_names Tx=Tx,Row=Row,Rows=Rows github.com/jackc/pgx/v5 Tx,Row,Rows

## test: Shortcut to launch all the test tasks (unit, functional and integration).
.PHONY: test
test:
	scripts/coverage
	$(MAKE) bench

## bench: Shortcut to launch benchmark tests.
.PHONY: bench
bench:
	go test $(PACKAGES) -bench . -benchmem -run Benchmark.*

## build: Build the application.
.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo $(PACKAGES)
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/$(APP_NAME) $(MAIN_SOURCE)
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/alcotest cmd/alcotest/alcotest.go

## config: Create local configuration
.PHONY: config
config:
	@cp .env.example .env

## run: Locally run the application, e.g. node index.js, python -m myapp, go run myapp etc ...
.PHONY: run
run:
	$(MAIN_RUNNER) \
		-rendererPathPrefix /app
