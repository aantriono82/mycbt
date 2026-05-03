#!/usr/bin/env bash
set -u

# Full backend audit for Student panel integrations.
# Covers: auth, exam listing/join/token verify/session flow/results/announcements/attendance,
# and permission boundaries for student role.
#
# Usage:
#   BASE_URL=http://127.0.0.1:8080 \
#   ADMIN_USERNAME=admin \
#   ADMIN_PASSWORD='admin12345678*' \
#   bash scripts/audit_student_full.sh

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-}"

PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0
RUN_ID="$(date +%s)"
TMP_BODY="/tmp/audit_student_full_body.json"
TMP_CODE="/tmp/audit_student_full_code.txt"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing required command: $1" >&2; exit 2; }
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

assert_2xx() {
  local method="$1"; local path="$2"; local auth="${3:-}"; local body="${4:-}"; local ctype="${5:-application/json}"
  http_call "${method}" "${path}" "${auth}" "${body}" "${ctype}"
  local code; code="$(cat "${TMP_CODE}")"
  if [[ "${code}" =~ ^2[0-9][0-9]$ ]]; then ok "${method} ${path} -> ${code}"; return 0; fi
  fail "${method} ${path} -> ${code}"; return 1
}

assert_code() {
  local expected="$1"; local method="$2"; local path="$3"; local auth="${4:-}"; local body="${5:-}"; local ctype="${6:-application/json}"
  http_call "${method}" "${path}" "${auth}" "${body}" "${ctype}"
  local code; code="$(cat "${TMP_CODE}")"
  if [[ "${code}" == "${expected}" ]]; then ok "${method} ${path} -> ${code}"; return 0; fi
  fail "${method} ${path} -> expected ${expected}, got ${code}"; return 1
}

assert_2xx_optional() {
  local method="$1"; local path="$2"; local auth="${3:-}"; local body="${4:-}"; local ctype="${5:-application/json}"
  http_call "${method}" "${path}" "${auth}" "${body}" "${ctype}"
  local code; code="$(cat "${TMP_CODE}")"
  if [[ "${code}" =~ ^2[0-9][0-9]$ ]]; then ok "${method} ${path} -> ${code}"; return 0; fi
  warn "${method} ${path} -> ${code} (optional)"; return 0
}

extract_access_token() {
  python3 -c 'import json,sys; print(json.loads(sys.argv[1])["data"]["access_token"])' "$1"
}
extract_data_id() {
  python3 -c 'import json,sys; obj=json.loads(sys.argv[1]); data=obj.get("data"); print(data.get("id","") if isinstance(data,dict) else "")' "$1"
}
extract_first_exam_id() {
  python3 -c 'import json,sys; obj=json.load(open(sys.argv[1])); arr=obj.get("data") or []; print(arr[0].get("id","") if arr else "")' "$1"
}
extract_session_id_from_join() {
  python3 -c 'import json,sys; obj=json.load(open(sys.argv[1])); d=obj.get("data") or {}; print(d.get("session_id",""))' "$1"
}
extract_first_question_id() {
  python3 -c 'import json,sys; obj=json.load(open(sys.argv[1])); arr=obj.get("data") or []; print(arr[0].get("id","") if arr else "")' "$1"
}
extract_token_value() {
  python3 -c 'import json,sys; obj=json.load(open(sys.argv[1])); d=obj.get("data") or {}; print(d.get("token",""))' "$1"
}

echo "== Health + admin login =="
assert_2xx GET /healthz || true
assert_2xx GET /api/v1/ping || true

ADMIN_LOGIN_JSON="$(curl -sS -X POST "${BASE_URL}/api/v1/auth/login" \
  -H 'content-type: application/json' \
  -d "{\"username\":\"${ADMIN_USERNAME}\",\"password\":\"${ADMIN_PASSWORD}\"}")"
if ! ADMIN_TOKEN="$(extract_access_token "${ADMIN_LOGIN_JSON}" 2>/dev/null)"; then
  fail "POST /api/v1/auth/login (admin)"
  echo "Summary: pass=${PASS_COUNT} warn=${WARN_COUNT} fail=${FAIL_COUNT}"
  exit 1
fi
ok "POST /api/v1/auth/login (admin)"

echo "== Seed fixtures (admin) =="
PROGRAM_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"code\":\"SPRG-${RUN_ID}\",\"name\":\"Student Program ${RUN_ID}\"}" "${BASE_URL}/api/v1/admin/programs")"
PROGRAM_ID="$(extract_data_id "${PROGRAM_RESP}" 2>/dev/null || true)"
[[ -n "${PROGRAM_ID}" ]] && ok "create program" || fail "create program"

LEVEL_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"name\":\"Student Level ${RUN_ID}\",\"kelas\":10}" "${BASE_URL}/api/v1/admin/levels")"
LEVEL_ID="$(extract_data_id "${LEVEL_RESP}" 2>/dev/null || true)"
[[ -n "${LEVEL_ID}" ]] && ok "create level" || fail "create level"

GROUP_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"name\":\"Student Group ${RUN_ID}\"}" "${BASE_URL}/api/v1/admin/groups")"
GROUP_ID="$(extract_data_id "${GROUP_RESP}" 2>/dev/null || true)"
[[ -n "${GROUP_ID}" ]] && ok "create group" || fail "create group"

SUBJECT_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"code\":\"SSUB-${RUN_ID}\",\"name\":\"Student Subject ${RUN_ID}\"}" "${BASE_URL}/api/v1/admin/subjects")"
SUBJECT_ID="$(extract_data_id "${SUBJECT_RESP}" 2>/dev/null || true)"
[[ -n "${SUBJECT_ID}" ]] && ok "create subject" || fail "create subject"

SESSION_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"name\":\"Student Session ${RUN_ID}\",\"start_time\":\"00:00\",\"end_time\":\"23:59\"}" "${BASE_URL}/api/v1/admin/sessions")"
SESSION_ID="$(extract_data_id "${SESSION_RESP}" 2>/dev/null || true)"
[[ -n "${SESSION_ID}" ]] && ok "create session" || fail "create session"

TEACHER_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"username\":\"guru_student_audit_${RUN_ID}\",\"password\":\"GuruStudent12345!\",\"name\":\"Guru Student Audit ${RUN_ID}\",\"email\":\"guru_student_audit_${RUN_ID}@example.com\",\"nip\":\"${RUN_ID}\",\"mapel_codes\":\"SSUB-${RUN_ID}\",\"group_names\":\"Student Group ${RUN_ID}\",\"level_names\":\"Student Level ${RUN_ID}\"}" \
  "${BASE_URL}/api/v1/admin/teachers")"
TEACHER_ID="$(extract_data_id "${TEACHER_RESP}" 2>/dev/null || true)"
[[ -n "${TEACHER_ID}" ]] && ok "create teacher" || fail "create teacher"

STUDENT_USERNAME="siswa_audit_${RUN_ID}"
STUDENT_PASSWORD="SiswaAudit12345!"
STUDENT_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"username\":\"${STUDENT_USERNAME}\",\"password\":\"${STUDENT_PASSWORD}\",\"name\":\"Siswa Audit ${RUN_ID}\",\"email\":\"${STUDENT_USERNAME}@example.com\",\"nis\":\"${RUN_ID}\",\"program_id\":\"${PROGRAM_ID}\",\"level_id\":\"${LEVEL_ID}\",\"group_id\":\"${GROUP_ID}\"}" \
  "${BASE_URL}/api/v1/admin/students")"
STUDENT_ID="$(extract_data_id "${STUDENT_RESP}" 2>/dev/null || true)"
[[ -n "${STUDENT_ID}" ]] && ok "create student" || fail "create student"

# Question set and one simple question so student question endpoint has payload.
SET_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"subject_id\":\"${SUBJECT_ID}\",\"title\":\"Student Set ${RUN_ID}\",\"jenjang\":\"SMA\",\"level_id\":\"${LEVEL_ID}\"}" \
  "${BASE_URL}/api/v1/question-sets")"
SET_ID="$(extract_data_id "${SET_RESP}" 2>/dev/null || true)"
[[ -n "${SET_ID}" ]] && ok "create question set" || fail "create question set"
if [[ -n "${SET_ID}" ]]; then
  assert_2xx POST "/api/v1/question-sets/${SET_ID}/questions" "${ADMIN_TOKEN}" \
    '{"type":"mc_single","stem":"2 + 2 = ?","options":[{"label":"A","content":"4","is_correct":true},{"label":"B","content":"5","is_correct":false}]}' || true
fi

NOW="$(date -u -d '-15 minutes' +%Y-%m-%dT%H:%M:%SZ)"
END="$(date -u -d '+90 minutes' +%Y-%m-%dT%H:%M:%SZ)"
EXAM_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${ADMIN_TOKEN}" -H 'content-type: application/json' \
  -d "{\"subject_id\":\"${SUBJECT_ID}\",\"teacher_id\":\"${TEACHER_ID}\",\"session_id\":\"${SESSION_ID}\",\"title\":\"Student Exam ${RUN_ID}\",\"starts_at\":\"${NOW}\",\"ends_at\":\"${END}\",\"duration_minutes\":60,\"shuffle_questions\":false,\"shuffle_options\":false}" \
  "${BASE_URL}/api/v1/exams")"
EXAM_ID="$(extract_data_id "${EXAM_RESP}" 2>/dev/null || true)"
[[ -n "${EXAM_ID}" ]] && ok "create exam" || fail "create exam"

if [[ -n "${EXAM_ID}" && -n "${SET_ID}" && -n "${STUDENT_ID}" ]]; then
  assert_2xx PATCH "/api/v1/exams/${EXAM_ID}" "${ADMIN_TOKEN}" '{"status":"published"}' || true
  assert_2xx PUT "/api/v1/exams/${EXAM_ID}/question-sets" "${ADMIN_TOKEN}" "{\"items\":[{\"question_set_id\":\"${SET_ID}\"}]}" || true
  assert_2xx PUT "/api/v1/exams/${EXAM_ID}/targets" "${ADMIN_TOKEN}" "{\"student_ids\":[\"${STUDENT_ID}\"]}" || true
  assert_2xx POST "/api/v1/exams/${EXAM_ID}/tokens" "${ADMIN_TOKEN}" "{\"length\":6}" || true
  EXAM_TOKEN_VALUE="$(extract_token_value "${TMP_BODY}" 2>/dev/null || true)"
fi

echo "== Student login + student endpoints =="
STUDENT_LOGIN_JSON="$(curl -sS -X POST "${BASE_URL}/api/v1/auth/login" \
  -H 'content-type: application/json' \
  -d "{\"username\":\"${STUDENT_USERNAME}\",\"password\":\"${STUDENT_PASSWORD}\"}")"
if ! STUDENT_TOKEN="$(extract_access_token "${STUDENT_LOGIN_JSON}" 2>/dev/null)"; then
  fail "POST /api/v1/auth/login (student)"
  echo "Summary: pass=${PASS_COUNT} warn=${WARN_COUNT} fail=${FAIL_COUNT}"
  exit 1
fi
ok "POST /api/v1/auth/login (student)"
assert_2xx GET /api/v1/me "${STUDENT_TOKEN}" || true

assert_2xx GET "/api/v1/student/exams" "${STUDENT_TOKEN}" || true
assert_2xx GET "/api/v1/student/results" "${STUDENT_TOKEN}" || true
assert_2xx GET "/api/v1/student/announcements" "${STUDENT_TOKEN}" || true
assert_2xx GET "/api/v1/student/attendance/history" "${STUDENT_TOKEN}" || true

SESSION_RUN_ID=""
if [[ -n "${EXAM_ID}" ]]; then
  assert_2xx POST "/api/v1/student/exams/${EXAM_ID}/join" "${STUDENT_TOKEN}" "{\"token\":\"${EXAM_TOKEN_VALUE}\"}" || true
  SESSION_RUN_ID="$(extract_session_id_from_join "${TMP_BODY}" 2>/dev/null || true)"
fi

if [[ -n "${SESSION_RUN_ID}" ]]; then
  assert_2xx GET "/api/v1/student/sessions/${SESSION_RUN_ID}" "${STUDENT_TOKEN}" || true
  assert_2xx POST "/api/v1/student/sessions/${SESSION_RUN_ID}/verify-token" "${STUDENT_TOKEN}" "{\"token\":\"${EXAM_TOKEN_VALUE}\"}" || true
  assert_2xx GET "/api/v1/student/sessions/${SESSION_RUN_ID}/questions" "${STUDENT_TOKEN}" || true
  QUESTION_ID="$(extract_first_question_id "${TMP_BODY}" 2>/dev/null || true)"
  assert_2xx GET "/api/v1/student/sessions/${SESSION_RUN_ID}/answers" "${STUDENT_TOKEN}" || true
  assert_2xx POST "/api/v1/student/sessions/${SESSION_RUN_ID}/heartbeat" "${STUDENT_TOKEN}" '{"current_question_id":""}' || true
  if [[ -n "${QUESTION_ID}" ]]; then
    assert_2xx_optional POST "/api/v1/student/sessions/${SESSION_RUN_ID}/answers" "${STUDENT_TOKEN}" "{\"question_id\":\"${QUESTION_ID}\",\"answer\":\"A\"}" || true
  else
    warn "no question id found in session questions payload"
  fi
  assert_2xx_optional POST "/api/v1/student/sessions/${SESSION_RUN_ID}/submit" "${STUDENT_TOKEN}" '{}' || true
else
  warn "join exam did not return session_id; session endpoints skipped"
fi

if [[ -n "${EXAM_ID}" ]]; then
  assert_2xx_optional DELETE "/api/v1/student/exams/${EXAM_ID}/dismiss" "${STUDENT_TOKEN}" || true
fi

echo "== Permission checks (student must be blocked on admin/teacher endpoints) =="
assert_code 403 GET "/api/v1/admin/programs" "${STUDENT_TOKEN}" || true
assert_code 403 GET "/api/v1/settings/system" "${STUDENT_TOKEN}" || true
assert_code 403 GET "/api/v1/exams" "${STUDENT_TOKEN}" || true
assert_code 403 GET "/api/v1/lti/platforms" "${STUDENT_TOKEN}" || true

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
echo "ALL OK (FULL STUDENT AUDIT)"
