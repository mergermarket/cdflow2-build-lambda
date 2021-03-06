FROM golang:alpine AS builder

WORKDIR /build

ADD go.mod go.sum ./
RUN go mod download
ADD . .
ENV CGO_ENABLED 0
RUN TEST_NO_DOCKER=true go test ./internal/app
RUN go build ./cmd/app

FROM alpine:latest

RUN apk add --no-cache docker-cli

COPY --from=builder /build/app /app

LABEL type="platform"

ENTRYPOINT ["/app"]
