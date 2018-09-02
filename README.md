# httputils

[![Build Status](https://travis-ci.org/ViBiOh/httputils.svg?branch=master)](https://travis-ci.org/ViBiOh/httputils) [![codecov](https://codecov.io/gh/ViBiOh/httputils/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/httputils) [![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/httputils)](https://goreportcard.com/report/github.com/ViBiOh/httputils)

Golang HTTP Utils func

## Alcotest

Make your server blow into balloon for checking its health.

### Usage

```bash
Usage of alcotest:
  -url string
      [health] URL to check
```

in Dockerfile

```bash
HEALTHCHECK --retries=10 CMD /alcotest -url http://localhost/health

COPY alcotest /alcotest
```

### Why `alcotest` ?

Because main use is to healthcheck.
Because it's a partial anagram of "ch**e**ck **stat**us **co**de ur**l**"

## Certificate generator

Generate a certificate for a web server.

### Usage

```bash
Usage of cmd/cert/cert.go:
  -cert string
      [tls] PEM Certificate file
  -hosts string
      [tls] Self-signed certificate hosts, comma separated (default "localhost")
  -key string
      [tls] PEM Key file
  -organization string
      [tls] Self-signed certificate organization (default "ViBiOh")
```
