default: go

go: deps dev

dev: format lint tst bench build

deps:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/golang/lint/golint
	go get -u golang.org/x/tools/cmd/goimports
	dep ensure

format:
	goimports -w **/*.go *.go
	gofmt -s -w **/*.go *.go

lint:
	golint `go list ./... | grep -v vendor`
	go vet ./...

tst:
	script/coverage

bench:
	go test ./... -bench . -benchmem -run Benchmark.*

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo cert/cert.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo cors/cors.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo db/db.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo owasp/owasp.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo prometheus/prometheus.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo rate/rate.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo tools/action.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo tools/map.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo tools/strings.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo writer/writer.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo pagination/pagination.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo breaksync/*.go
