#!/usr/bin/env bash
set -euo pipefail

# Teacher LTI E2E audit:
# - login admin
# - create teacher + subject + exam fixtures
# - create mock LTI platform + lti_session directly in DB
# - login teacher
# - call POST /api/v1/lti/deep-link (expect 200 + jwt)
# - cleanup
#
# Usage:
#   set -a; source backend/.env; set +a
#   ADMIN_USERNAME=admin ADMIN_PASSWORD='xxx' bash scripts/audit_teacher_lti_e2e.sh
#
# Required env:
#   DATABASE_URL must be set (for psql).

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-}"
DATABASE_URL="${DATABASE_URL:-}"

if [[ -z "${ADMIN_PASSWORD}" ]]; then
  echo "ADMIN_PASSWORD is required" >&2
  exit 2
fi

if [[ -z "${DATABASE_URL}" ]]; then
  echo "ERROR: DATABASE_URL is empty. Run: set -a; source backend/.env; set +a" >&2
  exit 2
fi

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing required command: $1" >&2; exit 2; }
}
need_cmd curl
need_cmd psql
need_cmd python3
need_cmd openssl

RUN_ID="$(date +%s)"
TMP="/tmp/audit_teacher_lti_e2e_body.json"

extract_access_token() {
  python3 -c 'import json,sys; print(json.loads(sys.argv[1])["data"]["access_token"])' "$1"
}

extract_data_id() {
  python3 -c 'import json,sys; obj=json.loads(sys.argv[1]); data=obj.get("data"); print(data.get("id","") if isinstance(data,dict) else "")' "$1"
}

psql_one() {
  local sql="$1"
  psql "${DATABASE_URL}" -qAtX -c "${sql}" | sed '/^\s*$/d' | head -n1 | tr -d '\r'
}

assert_2xx_json() {
  local method="$1"; local path="$2"; local token="$3"; local body="${4:-}"
  local code
  if [[ -n "${body}" ]]; then
    code="$(curl -sS -o "${TMP}" -w "%{http_code}" -X "${method}" \
      -H "Authorization: Bearer ${token}" \
      -H "content-type: application/json" \
      -d "${body}" \
      "${BASE_URL}${path}")"
  else
    code="$(curl -sS -o "${TMP}" -w "%{http_code}" -X "${method}" \
      -H "Authorization: Bearer ${token}" \
      "${BASE_URL}${path}")"
  fi
  if [[ ! "${code}" =~ ^2[0-9][0-9]$ ]]; then
    echo "FAIL ${method} ${path} -> ${code}" >&2
    head -c 500 "${TMP}" >&2 || true
    echo >&2
    exit 1
  fi
  echo "OK   ${method} ${path} -> ${code}"
}

echo "== Login admin =="
ADMIN_LOGIN_JSON="$(curl -sS -X POST "${BASE_URL}/api/v1/auth/login" \
  -H 'content-type: application/json' \
  -d "{\"username\":\"${ADMIN_USERNAME}\",\"password\":\"${ADMIN_PASSWORD}\"}")"
ADMIN_TOKEN="$(extract_access_token "${ADMIN_LOGIN_JSON}")"
[[ -n "${ADMIN_TOKEN}" ]] || { echo "FAIL admin login" >&2; exit 1; }
echo "OK   admin login"

echo "== Seed fixtures =="
SUBJECT_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"code\":\"LTI-SUB-${RUN_ID}\",\"name\":\"LTI Subject ${RUN_ID}\"}" \
  "${BASE_URL}/api/v1/admin/subjects")"
SUBJECT_ID="$(extract_data_id "${SUBJECT_RESP}")"

SESSION_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"name\":\"LTI Session ${RUN_ID}\",\"start_time\":\"07:30\",\"end_time\":\"09:00\"}" \
  "${BASE_URL}/api/v1/admin/sessions")"
SESSION_ID="$(extract_data_id "${SESSION_RESP}")"

TEACHER_USERNAME="guru_lti_${RUN_ID}"
TEACHER_PASSWORD="GuruLTI12345!"
TEACHER_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"username\":\"${TEACHER_USERNAME}\",\"password\":\"${TEACHER_PASSWORD}\",\"name\":\"Guru LTI ${RUN_ID}\",\"email\":\"${TEACHER_USERNAME}@example.com\",\"nip\":\"${RUN_ID}\",\"mapel_codes\":\"LTI-SUB-${RUN_ID}\"}" \
  "${BASE_URL}/api/v1/admin/teachers")"
TEACHER_ID="$(extract_data_id "${TEACHER_RESP}")"

NOW="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
END="$(date -u -d '+2 hours' +%Y-%m-%dT%H:%M:%SZ)"
EXAM_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"subject_id\":\"${SUBJECT_ID}\",\"teacher_id\":\"${TEACHER_ID}\",\"session_id\":\"${SESSION_ID}\",\"title\":\"LTI Exam ${RUN_ID}\",\"starts_at\":\"${NOW}\",\"ends_at\":\"${END}\",\"duration_minutes\":60}" \
  "${BASE_URL}/api/v1/exams")"
EXAM_ID="$(extract_data_id "${EXAM_RESP}")"

[[ -n "${SUBJECT_ID}" && -n "${SESSION_ID}" && -n "${TEACHER_ID}" && -n "${EXAM_ID}" ]] || {
  echo "FAIL fixture setup" >&2
  exit 1
}
echo "OK   fixture setup"

echo "== Login teacher =="
TEACHER_LOGIN_JSON="$(curl -sS -X POST "${BASE_URL}/api/v1/auth/login" \
  -H 'content-type: application/json' \
  -d "{\"username\":\"${TEACHER_USERNAME}\",\"password\":\"${TEACHER_PASSWORD}\"}")"
TEACHER_TOKEN="$(extract_access_token "${TEACHER_LOGIN_JSON}")"
[[ -n "${TEACHER_TOKEN}" ]] || { echo "FAIL teacher login" >&2; exit 1; }
echo "OK   teacher login"

echo "== Insert mock LTI platform + session =="
KEY_DIR="$(mktemp -d)"
PRIV_KEY_FILE="${KEY_DIR}/lti_private.pem"
PUB_KEY_FILE="${KEY_DIR}/lti_public.pem"
# Force PKCS#1 format (BEGIN RSA PRIVATE KEY) because backend parses PKCS1.
if ! openssl genrsa -traditional -out "${PRIV_KEY_FILE}" 2048 >/dev/null 2>&1; then
  openssl genrsa -out "${PRIV_KEY_FILE}" 2048 >/dev/null 2>&1
  openssl rsa -in "${PRIV_KEY_FILE}" -out "${PRIV_KEY_FILE}.pkcs1" -traditional >/dev/null 2>&1
  mv "${PRIV_KEY_FILE}.pkcs1" "${PRIV_KEY_FILE}"
fi
openssl rsa -in "${PRIV_KEY_FILE}" -pubout -out "${PUB_KEY_FILE}" >/dev/null 2>&1
PRIV_KEY_CONTENT="$(cat "${PRIV_KEY_FILE}")"
PUB_KEY_CONTENT="$(cat "${PUB_KEY_FILE}")"

PLATFORM_ID="$(psql_one "
INSERT INTO lti_platforms
  (name, issuer, client_id, deployment_id, oidc_auth_url, oidc_token_url, jwks_url, tool_private_key, tool_public_key)
VALUES
  ('LTI Mock ${RUN_ID}', 'https://lms.example.com/${RUN_ID}', 'client-${RUN_ID}', 'dep-${RUN_ID}',
   'https://lms.example.com/auth', 'https://lms.example.com/token', 'https://lms.example.com/jwks',
   \$\$${PRIV_KEY_CONTENT}\$\$,
   \$\$${PUB_KEY_CONTENT}\$\$)
RETURNING id::text;
")"

TEACHER_USER_ID="$(psql_one "SELECT user_id::text FROM teachers WHERE id='${TEACHER_ID}'")"
LTI_SESSION_ID="$(psql_one "
INSERT INTO lti_sessions
  (platform_id, local_user_id, message_type, return_url, data, deployment_id, expires_at)
VALUES
  ('${PLATFORM_ID}', '${TEACHER_USER_ID}', 'LtiDeepLinkingRequest', 'https://lms.example.com/return',
   'opaque-${RUN_ID}', 'dep-${RUN_ID}', NOW() + interval '30 minutes')
RETURNING id::text;
")"
[[ -n "${PLATFORM_ID}" && -n "${TEACHER_USER_ID}" && -n "${LTI_SESSION_ID}" ]] || { echo "FAIL create mock LTI session" >&2; exit 1; }
echo "OK   mock LTI session created"

echo "== Test teacher deep-link =="
assert_2xx_json POST /api/v1/lti/deep-link "${TEACHER_TOKEN}" \
  "{\"session_id\":\"${LTI_SESSION_ID}\",\"exam_id\":\"${EXAM_ID}\",\"title\":\"DeepLink LTI Exam ${RUN_ID}\"}"
python3 - <<'PY' "${TMP}"
import json,sys
obj=json.load(open(sys.argv[1]))
jwt=obj.get("data",{}).get("jwt","")
ret=obj.get("data",{}).get("return_url","")
assert jwt and ret, "jwt/return_url missing"
print("OK   deep-link payload contains jwt and return_url")
PY

echo "== Cleanup =="
psql "${DATABASE_URL}" -c "DELETE FROM lti_sessions WHERE id='${LTI_SESSION_ID}'" >/dev/null
psql "${DATABASE_URL}" -c "DELETE FROM lti_platforms WHERE id='${PLATFORM_ID}'" >/dev/null
curl -sS -X DELETE -H "Authorization: Bearer ${ADMIN_TOKEN}" "${BASE_URL}/api/v1/exams/${EXAM_ID}" >/dev/null || true
curl -sS -X DELETE -H "Authorization: Bearer ${ADMIN_TOKEN}" "${BASE_URL}/api/v1/admin/teachers/${TEACHER_ID}" >/dev/null || true
curl -sS -X DELETE -H "Authorization: Bearer ${ADMIN_TOKEN}" "${BASE_URL}/api/v1/admin/sessions/${SESSION_ID}" >/dev/null || true
curl -sS -X DELETE -H "Authorization: Bearer ${ADMIN_TOKEN}" "${BASE_URL}/api/v1/admin/subjects/${SUBJECT_ID}" >/dev/null || true
rm -rf "${KEY_DIR}"
echo "OK   cleanup done"

echo "ALL OK (TEACHER LTI E2E AUDIT)"
