FROM golang:1.12 as builder

WORKDIR /app
COPY . .

RUN make

ARG CODECOV_TOKEN
RUN curl -s https://codecov.io/bash | bash
