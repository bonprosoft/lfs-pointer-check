FROM golang:1.18.2-alpine3.16 AS build

WORKDIR /src
COPY . .

ARG GOOS linux
ARG GOARCH amd64

ENV GOOS ${GOOS}
ENV GOARCH ${GOARCH}
ENV CGO_ENABLED=0

RUN mkdir -p /build
RUN go build -ldflags="-s -w -extldflags \"-static\"" -o /build/lfs-pointer-check main.go
