FROM golang:1.12 as builder

ARG CODECOV_TOKEN
WORKDIR /app
COPY . .

RUN make httputils \
 && curl -s https://codecov.io/bash | bash
