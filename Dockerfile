FROM golang:1.12 as builder

WORKDIR /app
COPY . .

RUN make httputils
