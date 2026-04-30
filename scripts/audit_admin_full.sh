#!/usr/bin/env bash
set -u

# Full backend audit for Admin panel integrations.
# Covers: auth, settings, analytics, lms, master data, question bank, exams,
# tokens, results/evaluation, announcements, registrations, logs, lti, maintenance.
#
# Usage:
#   BASE_URL=http://127.0.0.1:8080 \
#   ADMIN_USERNAME=admin \
#   ADMIN_PASSWORD='admin12345678*' \
#   bash scripts/audit_admin_full.sh

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin12345}"

PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0
RUN_ID="$(date +%s)"

TMP_BODY="/tmp/audit_admin_full_body.json"
TMP_CODE="/tmp/audit_admin_full_code.txt"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 2
  }
}

need_cmd curl
need_cmd python3

print_section() {
  echo
  echo "== $1 =="
}

ok() {
  PASS_COUNT=$((PASS_COUNT + 1))
  echo "OK   $1"
}

fail() {
  FAIL_COUNT=$((FAIL_COUNT + 1))
  echo "FAIL $1"
  if [[ -f "${TMP_BODY}" ]]; then
    head -c 500 "${TMP_BODY}" || true
    echo
  fi
}

warn() {
  WARN_COUNT=$((WARN_COUNT + 1))
  echo "WARN $1"
  if [[ -f "${TMP_BODY}" ]]; then
    head -c 300 "${TMP_BODY}" || true
    echo
  fi
}

http_call() {
  local method="$1"
  local path="$2"
  local auth="${3:-}"
  local body="${4:-}"
  local ctype="${5:-application/json}"

  if [[ -n "${body}" ]]; then
    if [[ -n "${auth}" ]]; then
      curl -sS -o "${TMP_BODY}" -w "%{http_code}" -X "${method}" \
        -H "Authorization: Bearer ${auth}" \
        -H "content-type: ${ctype}" \
        -d "${body}" \
        "${BASE_URL}${path}" >"${TMP_CODE}"
    else
      curl -sS -o "${TMP_BODY}" -w "%{http_code}" -X "${method}" \
        -H "content-type: ${ctype}" \
        -d "${body}" \
        "${BASE_URL}${path}" >"${TMP_CODE}"
    fi
  else
    if [[ -n "${auth}" ]]; then
      curl -sS -o "${TMP_BODY}" -w "%{http_code}" -X "${method}" \
        -H "Authorization: Bearer ${auth}" \
        "${BASE_URL}${path}" >"${TMP_CODE}"
    else
      curl -sS -o "${TMP_BODY}" -w "%{http_code}" -X "${method}" \
        "${BASE_URL}${path}" >"${TMP_CODE}"
    fi
  fi
}

assert_2xx() {
  local method="$1"
  local path="$2"
  local auth="${3:-}"
  local body="${4:-}"
  local ctype="${5:-application/json}"

  http_call "${method}" "${path}" "${auth}" "${body}" "${ctype}"
  local code
  code="$(cat "${TMP_CODE}")"
  if [[ "${code}" =~ ^2[0-9][0-9]$ ]]; then
    ok "${method} ${path} -> ${code}"
    return 0
  fi
  fail "${method} ${path} -> ${code}"
  return 1
}

assert_code() {
  local expected="$1"
  local method="$2"
  local path="$3"
  local auth="${4:-}"
  local body="${5:-}"
  local ctype="${6:-application/json}"

  http_call "${method}" "${path}" "${auth}" "${body}" "${ctype}"
  local code
  code="$(cat "${TMP_CODE}")"
  if [[ "${code}" == "${expected}" ]]; then
    ok "${method} ${path} -> ${code}"
    return 0
  fi
  fail "${method} ${path} -> expected ${expected}, got ${code}"
  return 1
}

assert_2xx_optional() {
  local method="$1"
  local path="$2"
  local auth="${3:-}"
  local body="${4:-}"
  local ctype="${5:-application/json}"
  http_call "${method}" "${path}" "${auth}" "${body}" "${ctype}"
  local code
  code="$(cat "${TMP_CODE}")"
  if [[ "${code}" =~ ^2[0-9][0-9]$ ]]; then
    ok "${method} ${path} -> ${code}"
    return 0
  fi
  warn "${method} ${path} -> ${code} (optional check)"
  return 0
}

extract_access_token() {
  python3 -c 'import json,sys; print(json.loads(sys.argv[1])["data"]["access_token"])' "$1"
}

extract_data_id() {
  python3 -c 'import json,sys; obj=json.loads(sys.argv[1]); data=obj.get("data"); print(data.get("id","") if isinstance(data,dict) else ""); sys.exit(0 if (isinstance(data,dict) and data.get("id")) else 1)' "$1"
}

json_get_key() {
  local json="$1"
  local key="$2"
  python3 -c 'import json,sys; obj=json.loads(sys.argv[1]); print(obj.get(sys.argv[2], ""))' "${json}" "${key}"
}

print_section "Health + Auth"
assert_2xx GET /healthz || true
assert_2xx GET /api/v1/ping || true

LOGIN_JSON="$(curl -sS -X POST "${BASE_URL}/api/v1/auth/login" \
  -H 'content-type: application/json' \
  -d "{\"username\":\"${ADMIN_USERNAME}\",\"password\":\"${ADMIN_PASSWORD}\"}")"

if ! TOKEN="$(extract_access_token "${LOGIN_JSON}" 2>/dev/null)"; then
  fail "POST /api/v1/auth/login -> invalid_credentials or malformed response"
  echo "Summary: pass=${PASS_COUNT} fail=${FAIL_COUNT}"
  exit 1
fi
ok "POST /api/v1/auth/login -> token issued"
assert_2xx GET /api/v1/me "${TOKEN}" || true

print_section "Admin Dashboard + Settings + Analytics + LMS"
assert_2xx GET /api/v1/admin/dashboard/stats "${TOKEN}" || true
assert_2xx GET /api/v1/settings/system "${TOKEN}" || true
assert_2xx GET /api/v1/settings/school-identity "${TOKEN}" || true
assert_2xx GET /api/v1/settings/smtp "${TOKEN}" || true
assert_2xx GET /api/v1/settings/whatsapp "${TOKEN}" || true
assert_2xx GET /api/v1/analytics/dashboard "${TOKEN}" || true
assert_2xx GET /api/v1/lms/summary "${TOKEN}" || true
assert_2xx GET /api/v1/lms/exams "${TOKEN}" || true

print_section "Master Data CRUD"
PROGRAM_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${TOKEN}" -H 'content-type: application/json' \
  -d "{\"code\":\"PRG-${RUN_ID}\",\"name\":\"Program ${RUN_ID}\"}" \
  "${BASE_URL}/api/v1/admin/programs")"
PROGRAM_ID="$(extract_data_id "${PROGRAM_RESP}" 2>/dev/null || true)"
[[ -n "${PROGRAM_ID}" ]] && ok "POST /api/v1/admin/programs" || fail "POST /api/v1/admin/programs"

LEVEL_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${TOKEN}" -H 'content-type: application/json' \
  -d "{\"name\":\"Level ${RUN_ID}\",\"kelas\":10}" \
  "${BASE_URL}/api/v1/admin/levels")"
LEVEL_ID="$(extract_data_id "${LEVEL_RESP}" 2>/dev/null || true)"
[[ -n "${LEVEL_ID}" ]] && ok "POST /api/v1/admin/levels" || fail "POST /api/v1/admin/levels"

GROUP_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${TOKEN}" -H 'content-type: application/json' \
  -d "{\"name\":\"Group ${RUN_ID}\"}" \
  "${BASE_URL}/api/v1/admin/groups")"
GROUP_ID="$(extract_data_id "${GROUP_RESP}" 2>/dev/null || true)"
[[ -n "${GROUP_ID}" ]] && ok "POST /api/v1/admin/groups" || fail "POST /api/v1/admin/groups"

SUBJECT_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${TOKEN}" -H 'content-type: application/json' \
  -d "{\"code\":\"MAT-${RUN_ID}\",\"name\":\"Matematika ${RUN_ID}\"}" \
  "${BASE_URL}/api/v1/admin/subjects")"
SUBJECT_ID="$(extract_data_id "${SUBJECT_RESP}" 2>/dev/null || true)"
[[ -n "${SUBJECT_ID}" ]] && ok "POST /api/v1/admin/subjects" || fail "POST /api/v1/admin/subjects"

SESSION_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${TOKEN}" -H 'content-type: application/json' \
  -d "{\"name\":\"Sesi ${RUN_ID}\",\"start_time\":\"07:30\",\"end_time\":\"09:00\"}" \
  "${BASE_URL}/api/v1/admin/sessions")"
SESSION_ID="$(extract_data_id "${SESSION_RESP}" 2>/dev/null || true)"
[[ -n "${SESSION_ID}" ]] && ok "POST /api/v1/admin/sessions" || fail "POST /api/v1/admin/sessions"

[[ -n "${PROGRAM_ID}" ]] && assert_2xx GET "/api/v1/admin/programs/${PROGRAM_ID}" "${TOKEN}" || true
[[ -n "${LEVEL_ID}" ]] && assert_2xx GET "/api/v1/admin/levels/${LEVEL_ID}" "${TOKEN}" || true
[[ -n "${GROUP_ID}" ]] && assert_2xx GET "/api/v1/admin/groups/${GROUP_ID}" "${TOKEN}" || true
[[ -n "${SUBJECT_ID}" ]] && assert_2xx GET "/api/v1/admin/subjects/${SUBJECT_ID}" "${TOKEN}" || true
[[ -n "${SESSION_ID}" ]] && assert_2xx GET "/api/v1/admin/sessions/${SESSION_ID}" "${TOKEN}" || true

print_section "Teacher + Student + Lookups"
TEACHER_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${TOKEN}" -H 'content-type: application/json' \
  -d "{\"username\":\"guru_${RUN_ID}\",\"password\":\"guru12345\",\"name\":\"Guru ${RUN_ID}\",\"email\":\"guru_${RUN_ID}@example.com\",\"nip\":\"${RUN_ID}\",\"mapel_codes\":\"MAT-${RUN_ID}\",\"group_names\":\"Group ${RUN_ID}\",\"level_names\":\"Level ${RUN_ID}\"}" \
  "${BASE_URL}/api/v1/admin/teachers")"
TEACHER_ID="$(extract_data_id "${TEACHER_RESP}" 2>/dev/null || true)"
[[ -n "${TEACHER_ID}" ]] && ok "POST /api/v1/admin/teachers" || fail "POST /api/v1/admin/teachers"

STUDENT_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${TOKEN}" -H 'content-type: application/json' \
  -d "{\"username\":\"siswa_${RUN_ID}\",\"password\":\"siswa12345\",\"name\":\"Siswa ${RUN_ID}\",\"email\":\"siswa_${RUN_ID}@example.com\",\"nis\":\"${RUN_ID}\",\"program_id\":\"${PROGRAM_ID}\",\"level_id\":\"${LEVEL_ID}\",\"group_id\":\"${GROUP_ID}\"}" \
  "${BASE_URL}/api/v1/admin/students")"
STUDENT_ID="$(extract_data_id "${STUDENT_RESP}" 2>/dev/null || true)"
[[ -n "${STUDENT_ID}" ]] && ok "POST /api/v1/admin/students" || fail "POST /api/v1/admin/students"

assert_2xx GET "/api/v1/lookups/subjects" "${TOKEN}" || true
assert_2xx GET "/api/v1/lookups/sessions" "${TOKEN}" || true
assert_2xx GET "/api/v1/lookups/teachers" "${TOKEN}" || true
assert_2xx GET "/api/v1/lookups/students" "${TOKEN}" || true

print_section "Question Bank + Exams + Tokens + Evaluation + Print dependencies"
SET_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${TOKEN}" -H 'content-type: application/json' \
  -d "{\"subject_id\":\"${SUBJECT_ID}\",\"title\":\"Set ${RUN_ID}\",\"jenjang\":\"SMA\",\"level_id\":\"${LEVEL_ID}\"}" \
  "${BASE_URL}/api/v1/question-sets")"
SET_ID="$(extract_data_id "${SET_RESP}" 2>/dev/null || true)"
[[ -n "${SET_ID}" ]] && ok "POST /api/v1/question-sets" || fail "POST /api/v1/question-sets"
[[ -n "${SET_ID}" ]] && assert_2xx GET "/api/v1/question-sets/${SET_ID}/questions" "${TOKEN}" || true

NOW="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
END="$(date -u -d '+2 hours' +%Y-%m-%dT%H:%M:%SZ)"
EXAM_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${TOKEN}" -H 'content-type: application/json' \
  -d "{\"subject_id\":\"${SUBJECT_ID}\",\"teacher_id\":\"${TEACHER_ID}\",\"session_id\":\"${SESSION_ID}\",\"title\":\"Ujian ${RUN_ID}\",\"starts_at\":\"${NOW}\",\"ends_at\":\"${END}\",\"duration_minutes\":60,\"shuffle_questions\":false,\"shuffle_options\":false}" \
  "${BASE_URL}/api/v1/exams")"
EXAM_ID="$(extract_data_id "${EXAM_RESP}" 2>/dev/null || true)"
[[ -n "${EXAM_ID}" ]] && ok "POST /api/v1/exams" || fail "POST /api/v1/exams"

if [[ -n "${EXAM_ID}" ]]; then
  assert_2xx PUT "/api/v1/exams/${EXAM_ID}/question-sets" "${TOKEN}" "{\"items\":[{\"question_set_id\":\"${SET_ID}\"}]}" || true
  assert_2xx PUT "/api/v1/exams/${EXAM_ID}/targets" "${TOKEN}" "{\"student_ids\":[\"${STUDENT_ID}\"]}" || true
  assert_2xx GET "/api/v1/exams/${EXAM_ID}/tokens" "${TOKEN}" || true
  TOKEN_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${TOKEN}" -H 'content-type: application/json' \
    -d "{\"token\":\"123456\",\"is_active\":true}" \
    "${BASE_URL}/api/v1/exams/${EXAM_ID}/tokens")"
  EXAM_TOKEN_ID="$(extract_data_id "${TOKEN_RESP}" 2>/dev/null || true)"
  [[ -n "${EXAM_TOKEN_ID}" ]] && ok "POST /api/v1/exams/${EXAM_ID}/tokens" || fail "POST /api/v1/exams/${EXAM_ID}/tokens"
  [[ -n "${EXAM_TOKEN_ID}" ]] && assert_2xx PATCH "/api/v1/tokens/${EXAM_TOKEN_ID}" "${TOKEN}" "{\"is_active\":false}" || true
  assert_2xx POST "/api/v1/exams/${EXAM_ID}/tokens/rotate" "${TOKEN}" "{}" || true

  assert_2xx GET "/api/v1/exams/${EXAM_ID}/results" "${TOKEN}" || true
  assert_2xx GET "/api/v1/exams/${EXAM_ID}/score-distribution" "${TOKEN}" || true
  assert_2xx GET "/api/v1/exams/${EXAM_ID}/item-analysis" "${TOKEN}" || true
  assert_2xx GET "/api/v1/exams/${EXAM_ID}/attendance" "${TOKEN}" || true
  assert_2xx GET "/api/v1/exams/${EXAM_ID}/export" "${TOKEN}" || true
  assert_2xx_optional GET "/api/v1/exams/${EXAM_ID}/export.pdf" "${TOKEN}" || true
fi

assert_2xx GET "/api/v1/admin/levels?limit=100&offset=0" "${TOKEN}" || true
assert_2xx GET "/api/v1/admin/groups?limit=100&offset=0" "${TOKEN}" || true
assert_2xx GET "/api/v1/admin/sessions?limit=100&offset=0" "${TOKEN}" || true
assert_2xx GET "/api/v1/exams?limit=200&offset=0" "${TOKEN}" || true

print_section "Announcements + Registrations + Logs + LTI"
ANN_RESP="$(curl -sS -X POST -H "Authorization: Bearer ${TOKEN}" -H 'content-type: application/json' \
  -d "{\"title\":\"Pengumuman ${RUN_ID}\",\"body\":\"Isi pengumuman test\",\"category\":\"pengumuman\",\"is_active\":true}" \
  "${BASE_URL}/api/v1/announcements")"
ANN_ID="$(extract_data_id "${ANN_RESP}" 2>/dev/null || true)"
[[ -n "${ANN_ID}" ]] && ok "POST /api/v1/announcements" || fail "POST /api/v1/announcements"
assert_2xx GET "/api/v1/announcements?limit=10&offset=0" "${TOKEN}" || true
[[ -n "${ANN_ID}" ]] && assert_2xx GET "/api/v1/announcements/${ANN_ID}" "${TOKEN}" || true

assert_2xx GET "/api/v1/admin/registrations?limit=10&offset=0" "${TOKEN}" || true
assert_2xx GET "/api/v1/admin/registrations/pending?limit=10&offset=0" "${TOKEN}" || true
assert_2xx GET "/api/v1/admin/login-logs?limit=10&offset=0" "${TOKEN}" || true
assert_2xx GET "/api/v1/admin/audit-logs?limit=10&offset=0" "${TOKEN}" || true
assert_2xx GET "/api/v1/admin/audit-logs/export?limit=20&offset=0" "${TOKEN}" || true
assert_2xx GET "/api/v1/lti/platforms" "${TOKEN}" || true
assert_2xx POST "/api/v1/lti/keys/generate" "${TOKEN}" "{}" || true

print_section "Maintenance (safe checks)"
assert_2xx GET "/api/v1/maintenance/backup" "${TOKEN}" || true
# restore without file should return 400/415; this verifies endpoint is reachable and validates input.
assert_code 400 POST "/api/v1/maintenance/restore" "${TOKEN}" "" "application/x-www-form-urlencoded" || true

print_section "Cleanup (best effort)"
[[ -n "${EXAM_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/exams/${EXAM_ID}" >/dev/null || true
[[ -n "${SET_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/question-sets/${SET_ID}" >/dev/null || true
[[ -n "${ANN_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/announcements/${ANN_ID}" >/dev/null || true
[[ -n "${TEACHER_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/admin/teachers/${TEACHER_ID}" >/dev/null || true
[[ -n "${STUDENT_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/admin/students/${STUDENT_ID}" >/dev/null || true
[[ -n "${SESSION_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/admin/sessions/${SESSION_ID}" >/dev/null || true
[[ -n "${SUBJECT_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/admin/subjects/${SUBJECT_ID}" >/dev/null || true
[[ -n "${GROUP_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/admin/groups/${GROUP_ID}" >/dev/null || true
[[ -n "${LEVEL_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/admin/levels/${LEVEL_ID}" >/dev/null || true
[[ -n "${PROGRAM_ID:-}" ]] && curl -sS -X DELETE -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/admin/programs/${PROGRAM_ID}" >/dev/null || true
ok "cleanup done"

echo
echo "Summary: pass=${PASS_COUNT} warn=${WARN_COUNT} fail=${FAIL_COUNT}"
if [[ "${FAIL_COUNT}" -gt 0 ]]; then
  exit 1
fi
echo "ALL OK (FULL ADMIN AUDIT)"
