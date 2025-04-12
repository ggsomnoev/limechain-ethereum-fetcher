FROM golang:1.24 AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /build

COPY go.mod go.sum ./

COPY vendor ./vendor
ENV GOFLAGS=-mod=vendor

COPY internal ./internal
COPY pkg ./pkg
COPY cmd ./cmd

RUN go build -o main ./cmd/ethfetcher

FROM alpine:latest

WORKDIR /app
COPY --from=builder /build/main .

CMD ["./main"]
