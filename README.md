# httputils

[![Build](https://github.com/ViBiOh/httputils/workflows/Build/badge.svg)](https://github.com/ViBiOh/httputils/actions)
[![codecov](https://codecov.io/gh/ViBiOh/httputils/branch/main/graph/badge.svg)](https://codecov.io/gh/ViBiOh/httputils)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/httputils)](https://goreportcard.com/report/github.com/ViBiOh/httputils)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=ViBiOh_httputils&metric=alert_status)](https://sonarcloud.io/dashboard?id=ViBiOh_httputils)

Golang HTTP Utils func

## Alcotest

Make your server blow into balloon for checking its health.

### Usage

```bash
Usage of alcotest:
  -url string
        [alcotest] URL to check {ALCOTEST_URL}
  -userAgent string
        [alcotest] User-Agent for check {ALCOTEST_USER_AGENT} (default "Alcotest")
```

in `Dockerfile`

```bash
HEALTHCHECK --retries=10 CMD /alcotest -url http://localhost/health

COPY bin/alcotest /alcotest
```

### Why `alcotest` ?

Because main use is to healthcheck.
Because it's a partial anagram of "ch**e**ck **stat**us **co**de ur**l**"
