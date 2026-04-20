FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN go build -o app ./cmd/api

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]
