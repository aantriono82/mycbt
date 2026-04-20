FROM golang:1.22-alpine AS builder

WORKDIR /app

# copy go mod dulu
COPY backend/go.mod backend/go.sum ./backend/

WORKDIR /app/backend

RUN go mod download

# copy semua source
COPY backend/ .

# build dari cmd/api
RUN go build -o app ./cmd/api

# ===== runtime =====
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/backend/app .

EXPOSE 8080

CMD ["./app"]
