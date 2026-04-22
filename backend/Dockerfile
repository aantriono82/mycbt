# Stage 1: build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app ./cmd/api

# Stage 2: run
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]
