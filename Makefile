default: deps fmt lint tst build

deps:
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/golang/lint/golint

fmt:
	goimports -w **/*.go *.go
	gofmt -s -w **/*.go *.go

lint:
	golint ./...
	go vet ./...

tst:
	script/coverage

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo owasp/owasp.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo gzip/gzip.go
