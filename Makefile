APP_NAME ?= httputils
VERSION ?= $(shell git log --pretty=format:'%h' -n 1)
AUTHOR ?= $(shell git log --pretty=format:'%an' -n 1)

MAKEFLAGS += --silent
GOBIN=bin
PID=/tmp/.$(APP_NAME).pid

help: Makefile
	@sed -n 's|^##||p' $< | column -t -s ':' | sed -e 's|^| |'

## docker: Build App in a Docker container, tagged with current version
docker:
	docker build -t vibioh/$(APP_NAME):$(VERSION) .

$(APP_NAME): deps go

go: format lint tst bench build

## name: Output app name
name:
	@echo -n $(APP_NAME)

## version: Output last commit sha1
version:
	@echo -n $(VERSION)

## author: Output last commit author
author:
	@python -c 'import sys; import urllib; sys.stdout.write(urllib.quote_plus(sys.argv[1]))' "$(AUTHOR)"

deps:
	go get github.com/golang/dep/cmd/dep
	go get github.com/golang/lint/golint
	go get github.com/kisielk/errcheck
	go get golang.org/x/tools/cmd/goimports
	dep ensure

format:
	goimports -w **/*.go
	gofmt -s -w **/*.go

lint:
	golint `go list ./... | grep -v vendor`
	errcheck -ignoretests `go list ./... | grep -v vendor`
	go vet ./...

tst:
	script/coverage

bench:
	go test ./... -bench . -benchmem -run Benchmark.*

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo ./...
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o $(GOBIN)/alcotest cmd/alcotest.go

.PHONY: docker $(APP_NAME) go name version author deps format lint tst bench build
