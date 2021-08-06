# httputils

[![Build](https://github.com/ViBiOh/httputils/workflows/Build/badge.svg)](https://github.com/ViBiOh/httputils/actions)
[![codecov](https://codecov.io/gh/ViBiOh/httputils/branch/main/graph/badge.svg)](https://codecov.io/gh/ViBiOh/httputils)
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

## Server

Server is provided in order to be compliant with [12 Factor](https://12factor.net).

### Middlewares

#### health

Handle healthcheck status of the server.

#### recoverer

Catch `panic`, respond `HTTP/500` and log the beginning of the stacktrace.

#### prometheus

Increase metrics about HTTP requests.

#### owasp / cors

Enforce security best practices for serving web content.

### Endpoints

- `GET /health`: healthcheck of server, always respond [`okStatus (default 204)`](#usage)
- `GET /ready`: checks external dependencies availability and then respond [`okStatus (default 204)`](#usage) or `503` during [`graceDuration`](#usage) when `SIGTERM` is received
- `GET /version`: value of `VERSION` environment variable
- `GET /metrics`: Prometheus metrics, on a dedicated port [`prometheusPort (default 9090)`](#usage)
