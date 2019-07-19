SHELL = /bin/sh

ifneq ("$(wildcard .env)","")
	include .env
	export
endif

APP_NAME = alcotest
PACKAGES ?= ./...
GO_FILES ?= */*.go */*/*.go

GOBIN=bin
BINARY_PATH=$(GOBIN)/$(APP_NAME)

APP_SOURCE = cmd/alcotest/alcotest.go

.DEFAULT_GOAL := app

## help: Display list of commands
.PHONY: help
help: Makefile
	@sed -n 's|^##||p' $< | column -t -s ':' | sed -e 's|^| |'

## name: Output app name
.PHONY: name
name:
	@echo -n $(APP_NAME)

## version: Output last commit sha1
.PHONY: version
version:
	@echo -n $(shell git rev-parse --short HEAD)

## app: Build app with dependencies download
.PHONY: app
app: deps go

## go: Build app
.PHONY: go
go: format lint test bench build

## deps: Download dependencies
.PHONY: deps
deps:
	go get github.com/kisielk/errcheck
	go get golang.org/x/lint/golint
	go get golang.org/x/tools/cmd/goimports

## format: Format code of app
.PHONY: format
format:
	goimports -w $(GO_FILES)
	gofmt -s -w $(GO_FILES)

## lint: Lint code of app
.PHONY: lint
lint:
	golint $(PACKAGES)
	errcheck -ignoretests $(PACKAGES)
	go vet $(PACKAGES)

## test: Test code of app with coverage
.PHONY: test
test:
	script/coverage

## bench: Benchmark code of app
.PHONY: bench
bench:
	go test $(PACKAGES) -bench . -benchmem -run Benchmark.*

## build: Build binary of app
.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo $(PACKAGES)
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o $(BINARY_PATH) $(APP_SOURCE)
