# httputils

[![Build Status](https://travis-ci.org/ViBiOh/httputils.svg?branch=master)](https://travis-ci.org/ViBiOh/httputils) [![codecov](https://codecov.io/gh/ViBiOh/httputils/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/httputils) [![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/httputils)](https://goreportcard.com/report/github.com/ViBiOh/httputils)

Golang HTTP Utils func

## Alcotest

Make your server blow into balloon for checking its health.

### Usage

```
Usage of alcotest:
  -c string
      [health] URL to check
```

### Why argument with name `-c` for setting `url` ?

Main goal of this program is to be use for `HEALTHCHECK` in a Docker container, which run command from `/bin/sh -c`. So copying this binary as `/bin/sh` allow to only set url as `CMD` (cf. Dockerfile).

### Why `alcotest` ?

Because main use is to healthcheck.
Because it's a partial anagram of "ch**e**ck **stat**us **co**de ur**l**"