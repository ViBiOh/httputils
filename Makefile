APP_NAME ?= httputils

GOBIN=bin
BINARY_PATH=$(GOBIN)/$(APP_NAME)
VERSION ?= $(shell git log --pretty=format:'%h' -n 1)
AUTHOR ?= $(shell git log --pretty=format:'%an' -n 1)

help: Makefile
	@sed -n 's|^##||p' $< | column -t -s ':' | sed -e 's|^| |'

## $(APP_NAME): Build app with dependencies download
$(APP_NAME): deps go

go: format lint tst bench build

## name: Output app name
name:
	@echo -n $(APP_NAME)

## dist: Output build output path
dist:
	@echo -n $(BINARY_PATH)

## version: Output last commit sha1
version:
	@echo -n $(VERSION)

## author: Output last commit author
author:
	@python -c 'import sys; import urllib; sys.stdout.write(urllib.quote_plus(sys.argv[1]))' "$(AUTHOR)"

## deps: Download dependencies
deps:
	go get github.com/golang/dep/cmd/dep
	go get github.com/kisielk/errcheck
	go get golang.org/x/lint/golint
	go get golang.org/x/tools/cmd/goimports
	dep ensure

## format: Format code of app
format:
	goimports -w **/*.go
	gofmt -s -w **/*.go

## lint: Lint code of app
lint:
	golint `go list ./... | grep -v vendor`
	errcheck -ignoretests `go list ./... | grep -v vendor`
	go vet ./...

## tst: Test code of app with coverage
tst:
	script/coverage

## bench: Benchmark code of app
bench:
	go test ./... -bench . -benchmem -run Benchmark.*

## build: Build binary of app
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo ./...
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o $(GOBIN)/alcotest cmd/alcotest/alcotest.go

.PHONY: help $(APP_NAME) go name dist version author deps format lint tst bench build
