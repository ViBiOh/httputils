# httputils

[![Build Status](https://travis-ci.org/ViBiOh/httputils.svg?branch=master)](https://travis-ci.org/ViBiOh/httputils)
[![codecov](https://codecov.io/gh/ViBiOh/httputils/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/httputils)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/httputils)](https://goreportcard.com/report/github.com/ViBiOh/httputils)
[![Dependabot Status](https://api.dependabot.com/badges/status?host=github&repo=ViBiOh/httputils)](https://dependabot.com)

Golang HTTP Utils func

## Alcotest

Make your server blow into balloon for checking its health.

### Usage

```bash
Usage of alcotest:
  -url string
        [alcotest] URL to check {ALCOTEST_URL}
  -userAgent string
        [alcotest] User-Agent for check {ALCOTEST_USER_AGENT} (default "Golang alcotest")
```

in `Dockerfile`

```bash
HEALTHCHECK --retries=10 CMD /alcotest -url http://localhost/health

COPY bin/alcotest /alcotest
```

### Why `alcotest` ?

Because main use is to healthcheck.
Because it's a partial anagram of "ch**e**ck **stat**us **co**de ur**l**"
