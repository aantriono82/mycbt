#!/usr/bin/env bash
set -u

# Full backend audit for Teacher panel integrations.
# Strategy:
# 1) Login as admin, create isolated fixtures (subject/level/group/session/teacher/student).
# 2) Login as teacher, test teacher-facing endpoints and permissions.
# 3) Validate key negative cases (teacher blocked from admin-only endpoints).
# 4) Cleanup all fixtures.
#
# Usage:
#   BASE_URL=http://127.0.0.1:8080 \
#   ADMIN_USERNAME=admin \
#   ADMIN_PASSWORD='admin12345678*' \
#   bash scripts/audit_teacher_full.sh

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-}"

PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0
RUN_ID="$(date +%s)"
TMP_BODY="/tmp/audit_teacher_full_body.json"
TMP_CODE="/tmp/audit_teacher_full_code.txt"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 2
  }
}

need_cmd curl
need_cmd python3

if [[ -z "${ADMIN_PASSWORD}" ]]; then
  echo "ADMIN_PASSWORD is required" >&2
  exit 2
fi

ok() { PASS_COUNT=$((PASS_COUNT + 1)); echo "OK   $1"; }
fail() {
  FAIL_COUNT=$((FAIL_COUNT + 1))
  echo "FAIL $1"
  [[ -f "${TMP_BODY}" ]] && { head -c 500 "${TMP_BODY}" || true; echo; }
}

warn() {
  WARN_COUNT=$((WARN_COUNT + 1))
  echo "WARN $1"
  [[ -f "${TMP_BODY}" ]] && { head -c 300 "${TMP_BODY}" || true; echo; }
}

http_call() {
  local method="$1"; local path="$2"; local auth="${3:-}"; local body="${4:-}"; local ctype="${5:-application/json}"
  if [[ -n "${body}" ]]; then
    if [[ -n "${auth}" ]]; then
      curl -sS -o "${TMP_BODY}" -w "%{http_code}" -X "${method}" \
        -H "Authorization: Bearer ${auth}" -H "content-type: ${ctype}" \
        -d "${body}" "${BASE_URL}${path}" >"${TMP_CODE}"
    else
      curl -sS -o "${TMP_BODY}" -w "%{http_code}" -X "${method}" \
        -H "content-type: ${ctype}" -d "${body}" "${BASE_URL}${path}" >"${TMP_CODE}"
    fi
  else
    if [[ -n "${auth}" ]]; then
      curl -sS -o "${TMP_BODY}" -w "%{http_code}" -X "${method}" \
        -H "Authorization: Bearer ${auth}" "${BASE_URL}${path}" >"${TMP_CODE}"
    else
      curl -sS -o "${TMP_BODY}" -w "%{http_code}" -X "${method}" \
        "${BASE_URL}${path}" >"${TMP_CODE}"
    fi
  fi
}

assert_code() {
  local expected="$1"; local method="$2"; local path="$3"; local auth="${4:-}"; local body="${5:-}"; local ctype="${6:-application/json}"
  http_call "${method}" "${path}" "${auth}" "${body}" "${ctype}"
  local code; code="$(cat "${TMP_CODE}")"
  if [[ "${code}" == "${expected}" ]]; then ok "${method} ${path} -> ${code}"; return 0; fi
  fail "${method} ${path} -> expected ${expected}, got ${code}"; return 1
}

assert_2xx() {
  local method="$1"; local path="$2"; local auth="${3:-}"; local body="${4:-}"; local ctype="${5:-application/json}"
  http_call "${method}" "${path}" "${auth}" "${body}" "${ctype}"
  local code; code="$(cat "${TMP_CODE}")"
  if [[ "${code}" =~ ^2[0-9][0-9]$ ]]; then ok "${method} ${path} -> ${code}"; return 0; fi
  fail "${method} ${path} -> ${code}"; return 1
}

assert_2xx_optional() {
  local method="$1"; local path="$2"; local auth="${3:-}"; local body="${4:-}"; local ctype="${5:-application/json}"
  http_call "${method}" "${path}" "${auth}" "${body}" "${ctype}"
  local code; code="$(cat "${TMP_CODE}")"
  if [[ "${code}" =~ ^2[0-9][0-9]$ ]]; then ok "${method} ${path} -> ${code}"; return 0; fi
  warn "${method} ${path} -> ${code} (optional)"
  return 0
}

extract_access_token() {
  python3 -c 'import json,sys; print(json.loads(sys.argv[1])["data"]["access_token"])' "$1"
}

extract_data_id() {
  python3 -c 'import json,sys; obj=json.loads(sys.argv[1]); data=obj.get("data"); print(data.get("id","") if isinstance(data,dict) else ""); sys.exit(0 if (isinstance(data,dict) and data.get("id")) else 1)' "$1"
}

echo "== Health + admin login =="
assert_2xx GET /healthz || true
assert_2xx GET /api/v1/ping || true

ADMIN_LOGIN_JSON="$(curl -sS -X POST "${BASE_URL}/api/v1/auth/login" \
  -H 'content-type: application/json' \
  -d "{\"username\":\"${ADMIN_USERNAME}\",\"password\":\"${ADMIN_PASSWORD}\"}")"
if ! ADMIN_TOKEN="$(extract_access_token "${ADMIN_LOGIN_JSON}" 2>/dev/null)"; then
  fail "POST /api/v1/auth/login (admin)"
  echo "Summary: pass=${PASS_COUNT} fail=${FAIL_COUNT}"
  exit 1
fi
ok "POST /api/v1/auth/login (admin)"

echo "== Seed fixtures via admin =="
PROGRAM_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"code\":\"TPRG-${RUN_ID}\",\"name\":\"Teacher Program ${RUN_ID}\"}" "${BASE_URL}/api/v1/admin/programs")"
PROGRAM_ID="$(extract_data_id "${PROGRAM_RESP}" 2>/dev/null || true)"
[[ -n "${PROGRAM_ID}" ]] && ok "create program" || fail "create program"

LEVEL_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"name\":\"Teacher Level ${RUN_ID}\",\"kelas\":10}" "${BASE_URL}/api/v1/admin/levels")"
LEVEL_ID="$(extract_data_id "${LEVEL_RESP}" 2>/dev/null || true)"
[[ -n "${LEVEL_ID}" ]] && ok "create level" || fail "create level"

GROUP_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"name\":\"Teacher Group ${RUN_ID}\"}" "${BASE_URL}/api/v1/admin/groups")"
GROUP_ID="$(extract_data_id "${GROUP_RESP}" 2>/dev/null || true)"
[[ -n "${GROUP_ID}" ]] && ok "create group" || fail "create group"

SUBJECT_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"code\":\"TSUB-${RUN_ID}\",\"name\":\"Teacher Subject ${RUN_ID}\"}" "${BASE_URL}/api/v1/admin/subjects")"
SUBJECT_ID="$(extract_data_id "${SUBJECT_RESP}" 2>/dev/null || true)"
[[ -n "${SUBJECT_ID}" ]] && ok "create subject" || fail "create subject"

SESSION_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"name\":\"Teacher Session ${RUN_ID}\",\"start_time\":\"07:30\",\"end_time\":\"09:00\"}" "${BASE_URL}/api/v1/admin/sessions")"
SESSION_ID="$(extract_data_id "${SESSION_RESP}" 2>/dev/null || true)"
[[ -n "${SESSION_ID}" ]] && ok "create session" || fail "create session"

TEACHER_USERNAME="guru_audit_${RUN_ID}"
TEACHER_PASSWORD="GuruAudit12345!"
TEACHER_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"username\":\"${TEACHER_USERNAME}\",\"password\":\"${TEACHER_PASSWORD}\",\"name\":\"Guru Audit ${RUN_ID}\",\"email\":\"${TEACHER_USERNAME}@example.com\",\"nip\":\"${RUN_ID}\",\"mapel_codes\":\"TSUB-${RUN_ID}\",\"group_names\":\"Teacher Group ${RUN_ID}\",\"level_names\":\"Teacher Level ${RUN_ID}\"}" \
  "${BASE_URL}/api/v1/admin/teachers")"
TEACHER_ID="$(extract_data_id "${TEACHER_RESP}" 2>/dev/null || true)"
[[ -n "${TEACHER_ID}" ]] && ok "create teacher" || fail "create teacher"

STUDENT_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"username\":\"siswa_audit_${RUN_ID}\",\"password\":\"SiswaAudit12345!\",\"name\":\"Siswa Audit ${RUN_ID}\",\"email\":\"siswa_audit_${RUN_ID}@example.com\",\"nis\":\"${RUN_ID}\",\"program_id\":\"${PROGRAM_ID}\",\"level_id\":\"${LEVEL_ID}\",\"group_id\":\"${GROUP_ID}\"}" \
  "${BASE_URL}/api/v1/admin/students")"
STUDENT_ID="$(extract_data_id "${STUDENT_RESP}" 2>/dev/null || true)"
[[ -n "${STUDENT_ID}" ]] && ok "create student" || fail "create student"

echo "== Teacher login and teacher panel endpoints =="
TEACHER_LOGIN_JSON="$(curl -sS -X POST "${BASE_URL}/api/v1/auth/login" \
  -H 'content-type: application/json' \
  -d "{\"username\":\"${TEACHER_USERNAME}\",\"password\":\"${TEACHER_PASSWORD}\"}")"
if ! TEACHER_TOKEN="$(extract_access_token "${TEACHER_LOGIN_JSON}" 2>/dev/null)"; then
  fail "POST /api/v1/auth/login (teacher)"
  echo "Summary: pass=${PASS_COUNT} fail=${FAIL_COUNT}"
  exit 1
fi
ok "POST /api/v1/auth/login (teacher)"
assert_2xx GET /api/v1/me "${TEACHER_TOKEN}" || true

# Lookups used by teacher forms.
assert_2xx GET /api/v1/lookups/subjects "${TEACHER_TOKEN}" || true
assert_2xx GET /api/v1/lookups/sessions "${TEACHER_TOKEN}" || true
assert_2xx GET /api/v1/lookups/students "${TEACHER_TOKEN}" || true
assert_2xx GET /api/v1/lookups/my-assignments "${TEACHER_TOKEN}" || true

# Question bank by teacher.
SET_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${TEACHER_TOKEN}" -H 'content-type: application/json' \
  -d "{\"subject_id\":\"${SUBJECT_ID}\",\"title\":\"Teacher Set ${RUN_ID}\",\"jenjang\":\"SMA\",\"level_id\":\"${LEVEL_ID}\"}" \
  "${BASE_URL}/api/v1/question-sets")"
SET_ID="$(extract_data_id "${SET_RESP}" 2>/dev/null || true)"
[[ -n "${SET_ID}" ]] && ok "teacher create question set" || fail "teacher create question set"
[[ -n "${SET_ID}" ]] && assert_2xx GET "/api/v1/question-sets/${SET_ID}" "${TEACHER_TOKEN}" || true
[[ -n "${SET_ID}" ]] && assert_2xx GET "/api/v1/question-sets/${SET_ID}/questions" "${TEACHER_TOKEN}" || true

# Teacher exam flow.
NOW="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
END="$(date -u -d '+2 hours' +%Y-%m-%dT%H:%M:%SZ)"
EXAM_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${TEACHER_TOKEN}" -H 'content-type: application/json' \
  -d "{\"subject_id\":\"${SUBJECT_ID}\",\"session_id\":\"${SESSION_ID}\",\"title\":\"Teacher Exam ${RUN_ID}\",\"starts_at\":\"${NOW}\",\"ends_at\":\"${END}\",\"duration_minutes\":60,\"shuffle_questions\":false,\"shuffle_options\":false}" \
  "${BASE_URL}/api/v1/exams")"
EXAM_ID="$(extract_data_id "${EXAM_RESP}" 2>/dev/null || true)"
[[ -n "${EXAM_ID}" ]] && ok "teacher create exam" || fail "teacher create exam"

if [[ -n "${EXAM_ID}" ]]; then
  assert_2xx GET "/api/v1/exams/${EXAM_ID}" "${TEACHER_TOKEN}" || true
  assert_2xx PUT "/api/v1/exams/${EXAM_ID}/question-sets" "${TEACHER_TOKEN}" "{\"items\":[{\"question_set_id\":\"${SET_ID}\"}]}" || true
  assert_2xx PUT "/api/v1/exams/${EXAM_ID}/targets" "${TEACHER_TOKEN}" "{\"student_ids\":[\"${STUDENT_ID}\"]}" || true
  assert_2xx GET "/api/v1/exams/${EXAM_ID}/tokens" "${TEACHER_TOKEN}" || true
  assert_2xx POST "/api/v1/exams/${EXAM_ID}/tokens/rotate" "${TEACHER_TOKEN}" "{}" || true

  assert_2xx GET "/api/v1/exams/${EXAM_ID}/results" "${TEACHER_TOKEN}" || true
  assert_2xx GET "/api/v1/exams/${EXAM_ID}/score-distribution" "${TEACHER_TOKEN}" || true
  assert_2xx GET "/api/v1/exams/${EXAM_ID}/item-analysis" "${TEACHER_TOKEN}" || true
  assert_2xx GET "/api/v1/exams/${EXAM_ID}/attendance" "${TEACHER_TOKEN}" || true
  assert_2xx GET "/api/v1/exams/${EXAM_ID}/export" "${TEACHER_TOKEN}" || true
  assert_2xx GET "/api/v1/exams/${EXAM_ID}/export.pdf" "${TEACHER_TOKEN}" || true

  # Monitor + reset flow
  assert_2xx GET "/api/v1/exams/${EXAM_ID}/monitor/sessions" "${TEACHER_TOKEN}" || true
  assert_2xx GET "/api/v1/exams/${EXAM_ID}/monitor/participants" "${TEACHER_TOKEN}" || true
fi

# Shared modules accessible to teacher.
assert_2xx GET "/api/v1/analytics/dashboard" "${TEACHER_TOKEN}" || true
assert_2xx GET "/api/v1/lms/summary" "${TEACHER_TOKEN}" || true
assert_2xx GET "/api/v1/lms/exams" "${TEACHER_TOKEN}" || true
assert_2xx GET "/api/v1/announcements?limit=10&offset=0" "${TEACHER_TOKEN}" || true

# LTI deep-link route for teacher.
# Deep-link requires active LTI launch session (session_id), so optional in generic local audit.
assert_2xx_optional POST "/api/v1/lti/deep-link" "${TEACHER_TOKEN}" "{\"session_id\":\"dummy-session\",\"exam_id\":\"${EXAM_ID:-dummy-exam}\",\"title\":\"DL Test\"}" || true

echo "== Permission checks (teacher must be blocked on admin endpoints) =="
assert_code 403 GET "/api/v1/settings/system" "${TEACHER_TOKEN}" || true
assert_code 403 GET "/api/v1/admin/programs" "${TEACHER_TOKEN}" || true
assert_code 403 GET "/api/v1/lti/platforms" "${TEACHER_TOKEN}" || true
assert_code 403 GET "/api/v1/maintenance/backup" "${TEACHER_TOKEN}" || true

echo "== Cleanup =="
[[ -n "${EXAM_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${ADMIN_TOKEN}" "${BASE_URL}/api/v1/exams/${EXAM_ID}" >/dev/null || true
[[ -n "${SET_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${ADMIN_TOKEN}" "${BASE_URL}/api/v1/question-sets/${SET_ID}" >/dev/null || true
[[ -n "${TEACHER_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${ADMIN_TOKEN}" "${BASE_URL}/api/v1/admin/teachers/${TEACHER_ID}" >/dev/null || true
[[ -n "${STUDENT_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${ADMIN_TOKEN}" "${BASE_URL}/api/v1/admin/students/${STUDENT_ID}" >/dev/null || true
[[ -n "${SESSION_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${ADMIN_TOKEN}" "${BASE_URL}/api/v1/admin/sessions/${SESSION_ID}" >/dev/null || true
[[ -n "${SUBJECT_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${ADMIN_TOKEN}" "${BASE_URL}/api/v1/admin/subjects/${SUBJECT_ID}" >/dev/null || true
[[ -n "${GROUP_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${ADMIN_TOKEN}" "${BASE_URL}/api/v1/admin/groups/${GROUP_ID}" >/dev/null || true
[[ -n "${LEVEL_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${ADMIN_TOKEN}" "${BASE_URL}/api/v1/admin/levels/${LEVEL_ID}" >/dev/null || true
[[ -n "${PROGRAM_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${ADMIN_TOKEN}" "${BASE_URL}/api/v1/admin/programs/${PROGRAM_ID}" >/dev/null || true
ok "cleanup done"

echo
echo "Summary: pass=${PASS_COUNT} warn=${WARN_COUNT} fail=${FAIL_COUNT}"
if [[ "${FAIL_COUNT}" -gt 0 ]]; then
  exit 1
fi
echo "ALL OK (FULL TEACHER AUDIT)"
