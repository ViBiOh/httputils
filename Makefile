default: deps fmt lint tst build

deps:
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/golang/lint/golint

fmt:
	goimports -w *.go
	gofmt -s -w *.go

lint:
	golint ./...
	go vet ./...

tst:
	script/coverage

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo
