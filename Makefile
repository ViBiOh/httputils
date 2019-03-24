APP_NAME ?= httputils
VERSION ?= $(shell git log --pretty=format:'%h' -n 1)
AUTHOR ?= $(shell git log --pretty=format:'%an' -n 1)
PACKAGES ?= ./...

GOBIN=bin
BINARY_PATH=$(GOBIN)/$(APP_NAME)

.PHONY: help
help: Makefile
	@sed -n 's|^##||p' $< | column -t -s ':' | sed -e 's|^| |'

## $(APP_NAME): Build app with dependencies download
.PHONY: $(APP_NAME)
$(APP_NAME): deps go

.PHONY: go
go: format lint tst bench build

## name: Output app name
.PHONY: name
name:
	@echo -n $(APP_NAME)

## dist: Output build output path
.PHONY: dist
dist:
	@echo -n $(BINARY_PATH)

## version: Output last commit sha1
.PHONY: version
version:
	@echo -n $(VERSION)

## author: Output last commit author
.PHONY: author
author:
	@python -c 'import sys; import urllib; sys.stdout.write(urllib.quote_plus(sys.argv[1]))' "$(AUTHOR)"

## deps: Download dependencies
.PHONY: deps
deps:
	go get github.com/golang/dep/cmd/dep
	go get github.com/kisielk/errcheck
	go get golang.org/x/lint/golint
	go get golang.org/x/tools/cmd/goimports
	dep ensure

## format: Format code of app
.PHONY: format
format:
	goimports -w **/*.go
	gofmt -s -w **/*.go

## lint: Lint code of app
.PHONY: lint
lint:
	golint `go list $(PACKAGES) | grep -v vendor`
	errcheck -ignoretests `go list $(PACKAGES) | grep -v vendor`
	go vet $(PACKAGES)

## tst: Test code of app with coverage
.PHONY: test
tst:
	script/coverage

## bench: Benchmark code of app
.PHONY: bench
bench:
	go test $(PACKAGES) -bench . -benchmem -run Benchmark.*

## build: Build binary of app
.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo ./...
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o $(GOBIN)/alcotest cmd/alcotest/alcotest.go
