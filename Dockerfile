FROM golang:1.11 as builder

ENV APP_NAME httputils
ENV WORKDIR ${GOPATH}/src/github.com/ViBiOh/httputils

WORKDIR ${WORKDIR}
COPY ./ ${WORKDIR}/

RUN make ${APP_NAME}