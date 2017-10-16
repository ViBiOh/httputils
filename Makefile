default: deps format lint tst build

deps:
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/golang/lint/golint
	go get -u github.com/prometheus/client_golang/prometheus
	go get -u github.com/prometheus/client_golang/prometheus/promhttp

format:
	goimports -w **/*.go *.go
	gofmt -s -w **/*.go *.go

lint:
	golint ./...
	go vet ./...

tst:
	script/coverage

bench:
	go test ./... -bench . -benchmem -run Benchmark.*

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo owasp/owasp.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo cors/cors.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo prometheus/prometheus.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo cert/cert.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo tools/action.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo tools/map.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo rate/rate.go
