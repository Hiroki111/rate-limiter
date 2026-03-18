FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o redis-limiter-redis ./cmd/rate-limiter-redis/main.go

FROM alpine:latest
COPY --from=builder /app/redis-limiter-redis .
ENTRYPOINT ["./redis-limiter-redis", "--redis-addr=redis:6379"]