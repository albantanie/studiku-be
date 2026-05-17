# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/server ./main.go

FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

COPY --from=builder /app/server /app/server
COPY --from=builder /app/scripts /app/scripts

ENV SERVER_PORT=8080 \
    DB_HOST=postgres \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_NAME=studi_ku \
    DB_SSLMODE=disable \
    ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000

EXPOSE 8080
USER app

CMD ["/app/server", "rest"]
