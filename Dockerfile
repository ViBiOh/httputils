FROM golang:1.14 as builder

WORKDIR /app
COPY . .

RUN make \
 && git diff -- *.go \
 && git diff --quiet -- *.go

ARG CODECOV_TOKEN
RUN curl -q -sSL --max-time 10 https://codecov.io/bash | bash
