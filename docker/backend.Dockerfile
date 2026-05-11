FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /app/backend
RUN go mod download

COPY backend/ .
ENV CGO_ENABLED=0 GOFLAGS=-trimpath
RUN go build -ldflags="-s -w" -o /out/api ./cmd/api
RUN go build -ldflags="-s -w" -o /out/migrate ./cmd/migrate
RUN go build -ldflags="-s -w" -o /out/seed ./cmd/seed
RUN go build -ldflags="-s -w" -o /out/cleanup ./cmd/cleanup

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

COPY --from=builder /out/api ./api
COPY --from=builder /out/migrate ./migrate
COPY --from=builder /out/seed ./seed
COPY --from=builder /out/cleanup ./cleanup
COPY --from=builder /app/backend/migrations ./migrations

RUN mkdir -p /app/uploads

ENV GIN_MODE=release
ENV UPLOAD_LOCAL_DIR=/app/uploads

EXPOSE 8080

CMD ["./api"]
