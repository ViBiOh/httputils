FROM golang:1.12 as builder

WORKDIR /app
COPY . .

RUN make httputils

ARG CODECOV_TOKEN
RUN curl -s https://codecov.io/bash | bash
