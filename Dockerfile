FROM golang:1.13 as builder

WORKDIR /app
COPY . .

RUN make

ARG CODECOV_TOKEN
RUN curl -q -sS https://codecov.io/bash | bash
