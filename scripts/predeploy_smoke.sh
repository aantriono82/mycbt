#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-}"

if [[ -z "${ADMIN_PASSWORD}" ]]; then
  echo "ADMIN_PASSWORD is required" >&2
  exit 1
fi

echo "== Admin smoke =="
BASE_URL="${BASE_URL}" ADMIN_USERNAME="${ADMIN_USERNAME}" ADMIN_PASSWORD="${ADMIN_PASSWORD}" ./scripts/audit_admin.sh

echo "== Teacher smoke =="
BASE_URL="${BASE_URL}" ADMIN_USERNAME="${ADMIN_USERNAME}" ADMIN_PASSWORD="${ADMIN_PASSWORD}" ./scripts/audit_teacher_full.sh

echo "== Student smoke =="
BASE_URL="${BASE_URL}" ADMIN_USERNAME="${ADMIN_USERNAME}" ADMIN_PASSWORD="${ADMIN_PASSWORD}" ./scripts/audit_student_full.sh

echo "ALL PREDEPLOY SMOKE CHECKS PASSED"
