#!/usr/bin/env bash
set -euo pipefail

# Smoke-test backend endpoints used by Admin panel.
# Assumes API is already running and DB is migrated/seeded.
#
# Usage:
#   BASE_URL=http://127.0.0.1:8080 ADMIN_USERNAME=admin ADMIN_PASSWORD=admin12345 ./scripts/audit_admin.sh

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin12345}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 2
  }
}

need_cmd curl
need_cmd python3

json_get() {
  local url="$1"
  local auth="${2:-}"
  if [[ -n "$auth" ]]; then
    curl -sS -H "Authorization: Bearer ${auth}" "${BASE_URL}${url}"
  else
    curl -sS "${BASE_URL}${url}"
  fi
}

json_req() {
  local method="$1"
  local url="$2"
  local auth="$3"
  local body="$4"
  curl -sS -X "${method}" \
    -H "Authorization: Bearer ${auth}" \
    -H "content-type: application/json" \
    -d "${body}" \
    "${BASE_URL}${url}"
}

extract_data_id() {
  python3 -c 'import json,sys; obj=json.loads(sys.argv[1]); data=obj.get("data"); print(data.get("id","") if isinstance(data,dict) else ""); sys.exit(0 if (isinstance(data,dict) and data.get("id")) else 1)' "$1"
}

extract_access_token() {
  python3 -c 'import json,sys; raw=sys.argv[1]; obj=json.loads(raw); print(obj["data"]["access_token"])' "$1"
}

assert_http_200ish() {
  local method="$1"
  local url="$2"
  local auth="${3:-}"
  local body="${4:-}"
  local code
  local out

  if [[ -n "$body" ]]; then
    out="$(curl -sS -o /tmp/audit_admin_resp.json -w "%{http_code}" -X "${method}" \
      -H "Authorization: Bearer ${auth}" \
      -H "content-type: application/json" \
      -d "${body}" \
      "${BASE_URL}${url}")"
  else
    if [[ -n "$auth" ]]; then
      out="$(curl -sS -o /tmp/audit_admin_resp.json -w "%{http_code}" -X "${method}" \
        -H "Authorization: Bearer ${auth}" \
        "${BASE_URL}${url}")"
    else
      out="$(curl -sS -o /tmp/audit_admin_resp.json -w "%{http_code}" -X "${method}" \
        "${BASE_URL}${url}")"
    fi
  fi

  code="${out}"
  if [[ ! "${code}" =~ ^2[0-9][0-9]$ ]]; then
    echo "FAIL ${method} ${url} -> HTTP ${code}" >&2
    if [[ -f /tmp/audit_admin_resp.json ]]; then
      echo "Response:" >&2
      cat /tmp/audit_admin_resp.json >&2 || true
    fi
    exit 1
  fi
  echo "OK   ${method} ${url} -> ${code}"
}

RUN_ID="$(date +%s)"

echo "== Health =="
assert_http_200ish GET /healthz
assert_http_200ish GET /api/v1/ping

echo "== Login (admin) =="
LOGIN_JSON="$(curl -sS -X POST "${BASE_URL}/api/v1/auth/login" \
  -H 'content-type: application/json' \
  -d "{\"username\":\"${ADMIN_USERNAME}\",\"password\":\"${ADMIN_PASSWORD}\"}")"
if [[ -z "${LOGIN_JSON}" ]]; then
  echo "FAIL login: empty response body" >&2
  exit 1
fi
if ! TOKEN="$(extract_access_token "${LOGIN_JSON}")"; then
  echo "FAIL login: could not parse token from response (first 300 chars):" >&2
  printf "%s" "${LOGIN_JSON}" | head -c 300 >&2 || true
  echo >&2
  exit 1
fi
[[ -n "${TOKEN}" ]] || { echo "failed to obtain token" >&2; exit 1; }
assert_http_200ish GET /api/v1/me "${TOKEN}"

echo "== Settings / Analytics / LMS (GET only) =="
assert_http_200ish GET /api/v1/settings/system "${TOKEN}"
assert_http_200ish GET /api/v1/settings/school-identity "${TOKEN}"
assert_http_200ish GET /api/v1/analytics/dashboard "${TOKEN}"
assert_http_200ish GET /api/v1/lms/summary "${TOKEN}"
assert_http_200ish GET /api/v1/lms/exams "${TOKEN}"

echo "== Master Data CRUD (program/level/group/subject/session) =="

# Program
PROGRAM_RESP="$(json_req POST /api/v1/admin/programs "${TOKEN}" "{\"code\":\"PRG-${RUN_ID}\",\"name\":\"Program ${RUN_ID}\"}")"
PROGRAM_ID="$(extract_data_id "${PROGRAM_RESP}")"
assert_http_200ish GET "/api/v1/admin/programs/${PROGRAM_ID}" "${TOKEN}"
assert_http_200ish PATCH "/api/v1/admin/programs/${PROGRAM_ID}" "${TOKEN}" "{\"code\":\"PRG-${RUN_ID}\",\"name\":\"Program ${RUN_ID} (updated)\"}"

# Level
LEVEL_RESP="$(json_req POST /api/v1/admin/levels "${TOKEN}" "{\"name\":\"Level ${RUN_ID}\",\"kelas\":10}")"
LEVEL_ID="$(extract_data_id "${LEVEL_RESP}")"
assert_http_200ish GET "/api/v1/admin/levels/${LEVEL_ID}" "${TOKEN}"

# Group
GROUP_RESP="$(json_req POST /api/v1/admin/groups "${TOKEN}" "{\"name\":\"Group ${RUN_ID}\"}")"
GROUP_ID="$(extract_data_id "${GROUP_RESP}")"
assert_http_200ish GET "/api/v1/admin/groups/${GROUP_ID}" "${TOKEN}"

# Subject
SUBJECT_RESP="$(json_req POST /api/v1/admin/subjects "${TOKEN}" "{\"code\":\"MAT-${RUN_ID}\",\"name\":\"Matematika ${RUN_ID}\"}")"
SUBJECT_ID="$(extract_data_id "${SUBJECT_RESP}")"
assert_http_200ish GET "/api/v1/admin/subjects/${SUBJECT_ID}" "${TOKEN}"

# Session
SESSION_RESP="$(json_req POST /api/v1/admin/sessions "${TOKEN}" "{\"name\":\"Sesi ${RUN_ID}\",\"start_time\":\"07:30\",\"end_time\":\"09:00\"}")"
SESSION_ID="$(extract_data_id "${SESSION_RESP}")"
assert_http_200ish GET "/api/v1/admin/sessions/${SESSION_ID}" "${TOKEN}"

echo "== Teacher + Student create/list =="

TEACHER_RESP="$(json_req POST /api/v1/admin/teachers "${TOKEN}" \
  "{\"username\":\"guru_${RUN_ID}\",\"password\":\"guru12345\",\"name\":\"Guru ${RUN_ID}\",\"email\":\"guru_${RUN_ID}@example.com\",\"phone\":\"\",\"nip\":\"${RUN_ID}\",\"jenjang\":\"SMA\",\"mapel_codes\":\"MAT-${RUN_ID}\",\"group_names\":\"Group ${RUN_ID}\",\"level_names\":\"Level ${RUN_ID}\"}")"
TEACHER_ID="$(extract_data_id "${TEACHER_RESP}")"
assert_http_200ish GET "/api/v1/admin/teachers/${TEACHER_ID}" "${TOKEN}"

STUDENT_RESP="$(json_req POST /api/v1/admin/students "${TOKEN}" \
  "{\"username\":\"siswa_${RUN_ID}\",\"password\":\"siswa12345\",\"name\":\"Siswa ${RUN_ID}\",\"email\":\"siswa_${RUN_ID}@example.com\",\"phone\":\"\",\"nis\":\"${RUN_ID}\",\"jenjang\":\"SMA\",\"program_id\":\"${PROGRAM_ID}\",\"level_id\":\"${LEVEL_ID}\",\"group_id\":\"${GROUP_ID}\"}")"
STUDENT_ID="$(extract_data_id "${STUDENT_RESP}")"
assert_http_200ish GET "/api/v1/admin/students/${STUDENT_ID}" "${TOKEN}"

assert_http_200ish GET "/api/v1/admin/teachers?limit=5&offset=0&q=guru_${RUN_ID}" "${TOKEN}"
assert_http_200ish GET "/api/v1/admin/students?limit=5&offset=0&q=siswa_${RUN_ID}" "${TOKEN}"

echo "== Bank Soal (create set) =="
SET_RESP="$(json_req POST /api/v1/question-sets "${TOKEN}" \
  "{\"subject_id\":\"${SUBJECT_ID}\",\"title\":\"Set ${RUN_ID}\",\"jenjang\":\"SMA\",\"level_id\":\"${LEVEL_ID}\"}")"
SET_ID="$(extract_data_id "${SET_RESP}")"
assert_http_200ish GET "/api/v1/question-sets/${SET_ID}" "${TOKEN}"
assert_http_200ish GET "/api/v1/question-sets/${SET_ID}/questions" "${TOKEN}"

echo "== Ujian (create exam + targets + question sets + tokens) =="

NOW="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
END="$(date -u -d '+2 hours' +%Y-%m-%dT%H:%M:%SZ)"
EXAM_RESP="$(json_req POST /api/v1/exams "${TOKEN}" \
  "{\"subject_id\":\"${SUBJECT_ID}\",\"teacher_id\":\"${TEACHER_ID}\",\"session_id\":\"${SESSION_ID}\",\"title\":\"Ujian ${RUN_ID}\",\"starts_at\":\"${NOW}\",\"ends_at\":\"${END}\",\"duration_minutes\":60,\"shuffle_questions\":false,\"shuffle_options\":false}")"
EXAM_ID="$(extract_data_id "${EXAM_RESP}")"
assert_http_200ish GET "/api/v1/exams/${EXAM_ID}" "${TOKEN}"
assert_http_200ish PUT "/api/v1/exams/${EXAM_ID}/question-sets" "${TOKEN}" "{\"items\":[{\"question_set_id\":\"${SET_ID}\"}]}"
assert_http_200ish PUT "/api/v1/exams/${EXAM_ID}/targets" "${TOKEN}" "{\"level_ids\":[],\"group_ids\":[],\"student_ids\":[\"${STUDENT_ID}\"]}"

assert_http_200ish GET "/api/v1/exams/${EXAM_ID}/tokens" "${TOKEN}"
TOKEN_RESP="$(json_req POST "/api/v1/exams/${EXAM_ID}/tokens" "${TOKEN}" "{\"token\":\"123456\",\"is_active\":true}")"
TOKEN_ID="$(extract_data_id "${TOKEN_RESP}")"
assert_http_200ish PATCH "/api/v1/tokens/${TOKEN_ID}" "${TOKEN}" "{\"is_active\":false}"
assert_http_200ish POST "/api/v1/exams/${EXAM_ID}/tokens/rotate" "${TOKEN}" "{}"

echo "== Logs endpoints (GET only) =="
assert_http_200ish GET "/api/v1/admin/login-logs?limit=5&offset=0" "${TOKEN}"
assert_http_200ish GET "/api/v1/admin/audit-logs?limit=5&offset=0" "${TOKEN}"

echo "== Cleanup (best effort) =="
curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/exams/${EXAM_ID}" >/dev/null || true
curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/question-sets/${SET_ID}" >/dev/null || true
curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/admin/teachers/${TEACHER_ID}" >/dev/null || true
curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/admin/students/${STUDENT_ID}" >/dev/null || true
curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/admin/sessions/${SESSION_ID}" >/dev/null || true
curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/admin/subjects/${SUBJECT_ID}" >/dev/null || true
curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/admin/groups/${GROUP_ID}" >/dev/null || true
curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/admin/levels/${LEVEL_ID}" >/dev/null || true
curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/admin/programs/${PROGRAM_ID}" >/dev/null || true

echo "ALL OK"
