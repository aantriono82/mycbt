#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
MODE="${1:-all}"
TMP_DIR="$ROOT_DIR/.tmp"
GO_CACHE_DIR="$TMP_DIR/go-build"
GO_TMP_DIR="$TMP_DIR/go-tmp"

if [[ -n "${GO_BIN:-}" ]]; then
  GO_CMD="$GO_BIN"
elif [[ -x "$ROOT_DIR/.tooling/go/bin/go" ]]; then
  GO_CMD="$ROOT_DIR/.tooling/go/bin/go"
else
  GO_CMD="go"
fi

mkdir -p "$GO_CACHE_DIR" "$GO_TMP_DIR"
export GOCACHE="${GOCACHE:-$GO_CACHE_DIR}"
export GOTMPDIR="${GOTMPDIR:-$GO_TMP_DIR}"

run_all() {
  cd "$BACKEND_DIR"
  "$GO_CMD" test ./...
}

run_integration() {
  if [[ -z "${TEST_DATABASE_URL:-}" ]]; then
    echo "TEST_DATABASE_URL is required for integration mode." >&2
    echo "Example: TEST_DATABASE_URL='postgres://user:pass@localhost:5432/db?sslmode=disable' ./scripts/test-backend.sh integration" >&2
    exit 1
  fi

  cd "$BACKEND_DIR"
  "$GO_CMD" test -v \
    ./internal/testutil/pgtest \
    ./internal/repo/auditrepo \
    ./internal/repo/examrepo \
    ./internal/repo/loginlogrepo \
    ./internal/repo/ltirepo \
    ./internal/repo/masterrepo \
    ./internal/repo/questionbankrepo \
    ./internal/repo/userrepo
}

case "$MODE" in
  all)
    run_all
    ;;
  integration)
    run_integration
    ;;
  *)
    echo "Usage: $0 [all|integration]" >&2
    exit 1
    ;;
esac
