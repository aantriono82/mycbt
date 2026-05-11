#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "[cleanup] root: ${ROOT_DIR}"

rm -rf "${ROOT_DIR}/.gocache" \
       "${ROOT_DIR}/.tmp" \
       "${ROOT_DIR}/frontend/node_modules" \
       "${ROOT_DIR}/frontend/dist" \
       "${ROOT_DIR}/frontend/playwright-report" \
       "${ROOT_DIR}/frontend/test-results"

rm -f "${ROOT_DIR}/api_bin" \
      "${ROOT_DIR}/backend/api" \
      "${ROOT_DIR}/backend/api_new" \
      "${ROOT_DIR}/backend/api_sec" \
      "${ROOT_DIR}/backend/migrate" \
      "${ROOT_DIR}/backend/seed" \
      "${ROOT_DIR}/backend/cleanup" \
      "${ROOT_DIR}/backend/run.log"

echo "[cleanup] done"
