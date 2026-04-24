#!/bin/bash

# Configuration
export DATABASE_URL="postgres://mycbt:mycbt@localhost:5433/mycbt?sslmode=disable"
export JWT_SECRET="7f59f6b9c9f2b8e8a8b8c8d8e8f808182838485868788898a8b8c8d8e8f8081"
export GIN_MODE=debug

echo "Starting Postgres via Docker..."
docker compose up -d

echo "Starting Backend API..."
cd backend
if [ -x ./api ]; then
  ./api &
else
  ../.tooling/go/bin/go run ./cmd/api &
fi
BACKEND_PID=$!

echo "Starting Frontend..."
cd ../frontend
npm run dev &
FRONTEND_PID=$!

echo "Systems are starting up!"
echo "Backend PID: $BACKEND_PID"
echo "Frontend PID: $FRONTEND_PID"

# Wait for termination
trap "kill $BACKEND_PID $FRONTEND_PID; exit" SIGINT SIGTERM
wait
