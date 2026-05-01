# Backend (Gin)

OpenAPI baseline spec: [docs/openapi.yaml](/home/aantriono/dev/atigacbt/docs/openapi.yaml)

## Quickstart

Install Go locally (repo ships a local toolchain installer; no sudo needed):

```bash
./scripts/install-go.sh
```

Start Postgres (dev):

```bash
cd ..
docker compose up -d
cd backend
```

Set env (example):

```bash
export DATABASE_URL="postgres://atigacbt:atigacbt@localhost:5433/atigacbt?sslmode=disable"
export JWT_SECRET="$(openssl rand -hex 32)"
```

Run migrations:

```bash
../.tooling/go/bin/go run ./cmd/migrate
```

Seed admin:

```bash
export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=admin12345
../.tooling/go/bin/go run ./cmd/seed
```

Run API:

```bash
../.tooling/go/bin/go run ./cmd/api
```

## Testing

Backend sekarang punya kombinasi unit test, handler/service test, dan integration test PostgreSQL untuk repo yang query-heavy.

Perintah paling sederhana dari root repo:

```bash
make test-backend
```

Atau langsung lewat script:

```bash
./scripts/test-backend.sh
```

Mode default menjalankan:

```bash
cd backend
go test ./...
```

Catatan:
- suite ini aman dijalankan tanpa database test tambahan
- integration test repo yang butuh PostgreSQL akan `skip` otomatis saat `TEST_DATABASE_URL` belum diset

Untuk menjalankan integration test PostgreSQL secara eksplisit:

```bash
export TEST_DATABASE_URL='postgres://user:pass@localhost:5432/dbname?sslmode=disable'
make test-backend-integration
```

Atau:

```bash
TEST_DATABASE_URL='postgres://user:pass@localhost:5432/dbname?sslmode=disable' ./scripts/test-backend.sh integration
```

CI yang menjalankan suite ini ada di:
- [backend-tests.yml](/home/aantriono/dev/atigacbt/.github/workflows/backend-tests.yml)
  - job `test-backend` untuk `make test-backend`
  - job `test-backend-integration` untuk `make test-backend-integration` dengan service PostgreSQL

Perilaku harness integration test:
- setiap test run membuat schema acak sendiri
- semua file migrasi `backend/migrations/*.up.sql` diterapkan ke schema itu
- schema dibersihkan otomatis saat test selesai
- database yang dipakai harus PostgreSQL yang bisa diakses oleh `TEST_DATABASE_URL`

Area backend yang sekarang sudah punya coverage bermakna:
- auth handler + auth service
- student exam + exam results handler
- scoring helper dan manual grading helper
- LTI service dan notification service
- object storage lokal, config helper, db helper, migrate helper, PDF helper, cache helper, params helper, pg error helper
- integration test repo untuk `auditrepo`, `examrepo`, `loginlogrepo`, `ltirepo`, `masterrepo`, `questionbankrepo`, dan `userrepo`

Area yang sengaja tidak diprioritaskan untuk unit test detail:
- `cmd/*` entrypoints
- `internal/model`
- package root yang sangat tipis atau hanya wiring
- script operasional di `scripts/`

## Cleanup Pre-Deploy (Hapus Data Uji)

Sebelum go-live, gunakan cleanup agar akun guru/siswa dummy dan data ujian uji tidak ikut production.

SQL langsung:

```bash
psql "$DATABASE_URL" -f scripts/cleanup_predeploy.sql
```

CLI Go (lebih aman karena default `dry-run`):

```bash
# hanya lihat jumlah data (tanpa hapus)
go run ./cmd/cleanup

# eksekusi hapus data
go run ./cmd/cleanup --execute
```

Catatan:
- Script ini mempertahankan user role `admin`.
- Selalu backup database sebelum menjalankan mode `--execute`.

Endpoints (core):

- `GET /healthz`
- `GET /api/v1/ping`
- `POST /api/v1/auth/login`
- `GET /api/v1/me` (requires `Authorization: Bearer <token>`)

Ops tambahan yang sudah aktif:
- Header `X-Request-ID` otomatis pada semua response API.
- Structured logging JSON per request (`request_id`, status, latency, method/path, user role/id jika ada).
- Audit log mutasi (`POST/PUT/PATCH/DELETE`) role admin/guru disimpan ke tabel `audit_logs`.
- Unit test untuk parser soal, rate limiter middleware, dan evaluator scoring engine per tipe soal.
- Rate limit endpoint sensitif:
  - `POST /api/v1/auth/login`: 10 request/menit per IP
  - `POST /api/v1/student/exams/:id/join`: 20 request/menit per IP
- Upload size limit endpoint import:
  - Excel import guru/siswa: max 8 MB
  - DOCX import bank soal: max 24 MB (ditambah validasi file parser 20 MB)

Settings (Admin only; requires JWT):

- `GET /api/v1/settings/school-identity`
- `PUT /api/v1/settings/school-identity`
- `POST /api/v1/settings/school-identity/logo` (multipart `file`, png/jpg/jpeg/webp, max 4 MB)
- `GET /api/v1/settings/system`
- `PUT /api/v1/settings/system`
  - `token_required` diterapkan pada endpoint join siswa:
    - `true` -> token wajib di `POST /api/v1/student/exams/:id/join`
    - `false` -> token opsional (jika dikirim tetap divalidasi)
  - `max_active_sessions` diterapkan pada endpoint join siswa:
    - jika jumlah sesi `in_progress` siswa sudah mencapai batas, join baru ditolak `409 max_active_sessions_reached`
  - `allow_reset_login` diterapkan pada endpoint reset sesi:
    - `true` -> `POST /api/v1/exams/:id/sessions/:sessionId/reset` diizinkan
    - `false` -> endpoint reset mengembalikan `403 policy_disabled`
  - `attendance_require_ip` diterapkan pada endpoint absensi siswa:
    - `true` -> `POST /api/v1/student/attendance` mewajibkan `client_ip` valid dari request
    - data absensi menyimpan kolom `client_ip` (nullable)
  - Upload logo sekolah disimpan lokal di path `/uploads/logos/*` dan dapat diakses lewat static route `/uploads/*`

Master Data (Admin only; requires JWT):

- `GET /api/v1/admin/teachers`
- Supports query params: `q`, `limit`, `offset` (response includes `meta.total`)
- `POST /api/v1/admin/teachers` (manual create)
- `GET /api/v1/admin/teachers/:id`
- `GET /api/v1/admin/teachers/:id/subjects`
- `PUT /api/v1/admin/teachers/:id/subjects` (set mapel yang diampu)
- `PATCH /api/v1/admin/teachers/:id`
- `DELETE /api/v1/admin/teachers/:id`
- `GET /api/v1/admin/teachers/template` (xlsx)
- `POST /api/v1/admin/teachers/import` (multipart `file`)
- `GET /api/v1/admin/students`
- Supports query params: `q`, `limit`, `offset` (response includes `meta.total`)
- `POST /api/v1/admin/students` (manual create)
- `GET /api/v1/admin/students/:id`
- `PATCH /api/v1/admin/students/:id`
- `DELETE /api/v1/admin/students/:id`
- `GET /api/v1/admin/students/template` (xlsx)
- `POST /api/v1/admin/students/import` (multipart `file`)
- `GET /api/v1/admin/programs` / `POST /api/v1/admin/programs`
- `GET /api/v1/admin/programs/:id` / `PATCH` / `DELETE`
- `GET /api/v1/admin/levels` / `POST /api/v1/admin/levels`
- `GET /api/v1/admin/levels/:id` / `PATCH` / `DELETE`
- `GET /api/v1/admin/groups` / `POST /api/v1/admin/groups`
- `GET /api/v1/admin/groups/:id` / `PATCH` / `DELETE`
- `GET /api/v1/admin/subjects` / `POST /api/v1/admin/subjects`
- `GET /api/v1/admin/subjects/:id` / `PATCH` / `DELETE`
- `GET /api/v1/admin/announcements`
- `POST /api/v1/admin/announcements`
- `GET /api/v1/admin/announcements/:id`
- `PATCH /api/v1/admin/announcements/:id`
- `DELETE /api/v1/admin/announcements/:id`
- `GET /api/v1/admin/registrations/pending`
- `GET /api/v1/admin/registrations` (filters: `status`, `role`, `q`, `limit`, `offset`)
- `GET /api/v1/admin/registrations/:id`
- `PATCH /api/v1/admin/registrations/:id` (edit pending request)
- `POST /api/v1/admin/registrations/:id/approve`
- `POST /api/v1/admin/registrations/:id/reject`

Public registration (pending approval):

- `POST /api/v1/registrations`
- `GET /api/v1/registrations/:id` (check status)

Example login:

```bash
curl -sS -X POST http://localhost:8080/api/v1/auth/login \
  -H 'content-type: application/json' \
  -d '{"username":"admin","password":"admin12345"}'
```

Example student join exam (login -> list -> join with exam token):

```bash
export TOKEN="paste_access_token_di_sini"
curl -sS http://localhost:8080/api/v1/student/exams -H "Authorization: Bearer $TOKEN"

export EXAM_ID="paste_exam_id_di_sini"
curl -sS -X POST "http://localhost:8080/api/v1/student/exams/$EXAM_ID/join" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'content-type: application/json' \
  -d '{"token":"123456"}'
```

Example download template:

```bash
curl -fSLo template-guru.xlsx http://localhost:8080/api/v1/admin/teachers/template \
  -H "Authorization: Bearer $TOKEN"
```

Example import:

```bash
curl -fS http://localhost:8080/api/v1/admin/teachers/import \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@template-guru.xlsx"
```

## Bank Soal

Bank soal tersedia untuk role `admin` dan `teacher` (JWT).

Catatan role `teacher`:
- `subject_id` saat membuat bank soal wajib termasuk mapel yang diampu (di-set oleh admin lewat `PUT /api/v1/admin/teachers/:id/subjects`).

Endpoints:

- `GET/POST /api/v1/question-sets`
- `GET/PATCH/DELETE /api/v1/question-sets/:id`
- `GET/POST /api/v1/question-sets/:id/questions`
- `GET /api/v1/questions/:id`
- `PATCH /api/v1/questions/:id`
- `DELETE /api/v1/questions/:id`
- `POST /api/v1/question-sets/:id/import-docx/preview` (multipart `file`)
- `POST /api/v1/question-sets/:id/import-docx` (multipart `file`)

Payload `POST /question-sets/:id/questions` dan `PATCH /questions/:id` (6 tipe soal):

1) `mc_single` / `mc_multiple`

```json
{
  "type": "mc_single",
  "stem": "Ibu kota Indonesia adalah ...",
  "order_no": 1,
  "options": [
    {"label":"A","content":"Jakarta","is_correct":true},
    {"label":"B","content":"Bandung","is_correct":false}
  ]
}
```

2) `matching` (pairs minimal 2)

```json
{
  "type": "matching",
  "stem": "Jodohkan berikut:",
  "order_no": 1,
  "pairs": [
    {"left_content":"1","right_content":"Satu","order_no":1},
    {"left_content":"2","right_content":"Dua","order_no":2}
  ]
}
```

3) `short_answer` (answers minimal 1)

```json
{
  "type": "short_answer",
  "stem": "Sebutkan 1 contoh hewan mamalia",
  "order_no": 1,
  "answers": [
    {"answer_text":"kucing","order_no":1},
    {"answer_text":"sapi","order_no":2}
  ]
}
```

4) `true_false`

```json
{
  "type": "true_false",
  "stem": "2 + 2 = 4",
  "order_no": 1,
  "correct": true
}
```

5) `essay` (opsional rubric/max_score)

```json
{
  "type": "essay",
  "stem": "Jelaskan proses fotosintesis.",
  "order_no": 1,
  "rubric_text": "Nilai berdasarkan kelengkapan konsep.",
  "max_score": 100
}
```

## Ujian (Jadwal + Token)

Endpoints (admin/teacher; JWT):

- `GET/POST /api/v1/exams` (teacher: hanya ujian miliknya)
- `GET/PATCH/DELETE /api/v1/exams/:id`
- `GET/PUT /api/v1/exams/:id/targets`
- `GET/PUT /api/v1/exams/:id/question-sets`
- `GET/POST /api/v1/exams/:id/tokens`
- `PATCH /api/v1/tokens/:id` (set `is_active`)
- `GET /api/v1/exams/:id/results` (rekap nilai peserta)
- `GET /api/v1/exams/:id/item-analysis` (analisis butir: p-value, kategori kesukaran, daya pembeda upper/lower group, distraktor opsi)
- `GET /api/v1/exams/:id/score-distribution` (distribusi nilai + min/avg/median/max)
- `GET /api/v1/exams/:id/export` (export `.xlsx` multi-sheet: ExecutiveSummary + Results + Score Distribution + Item Analysis)
- `GET /api/v1/exams/:id/item-analysis/export` (export analisis butir + d-index + distraktor ke `.xlsx`)
- `GET /api/v1/exams/:id/attendance` (rekap absensi peserta + attendance rate)
- `GET /api/v1/exams/:id/monitor/sessions` (snapshot sesi peserta + progress; polling-friendly)
- `GET /api/v1/exams/:id/monitor/participants` (snapshot peserta target + status join/online/progress; polling-friendly)
- `GET /api/v1/exams/:id/monitor/stream?view=sessions|participants&access_token=...` (SSE stream untuk monitor live)
- `POST /api/v1/exams/:id/sessions/:sessionId/reset` (hapus session non-submitted agar siswa bisa join ulang)
- `POST /api/v1/exams/:id/sessions/:sessionId/force-submit` (paksa submit sesi `in_progress`; status akhir `forced`)

Catatan create exam:
- Role `teacher`: `teacher_id` otomatis dari akun.
- Role `teacher`: `subject_id` wajib termasuk mapel yang diampu (mapping `teacher_subjects`).
- Role `admin`: wajib kirim `teacher_id`.

Contoh create exam:

```json
{
  "subject_id": "uuid",
  "teacher_id": "uuid (admin only)",
  "title": "Ujian Harian 1",
  "starts_at": "2026-04-13T10:00:00Z",
  "ends_at": "2026-04-13T11:00:00Z",
  "duration_minutes": 60,
  "shuffle_questions": true,
  "shuffle_options": true
}
```

Contoh generate token:

```json
{
  "valid_from": "2026-04-13T09:55:00Z",
  "valid_to": "2026-04-13T11:05:00Z",
  "length": 6
}
```

Contoh set targets:

```json
{
  "level_ids": ["uuid"],
  "group_ids": ["uuid"],
  "student_ids": ["uuid"]
}
```

Contoh attach bank soal:

```json
{
  "items": [
    {"question_set_id": "uuid", "num_questions": 20}
  ]
}
```

## Siswa (Ruang Ujian)

Endpoints (student; JWT):

- `GET  /api/v1/student/exams` (ujian yang ditarget ke siswa; filter: `q`, `limit`, `offset`)
- `GET  /api/v1/student/exams/:examId/session` (cek sesi aktif untuk resume tanpa input token ulang)
- `POST /api/v1/student/exams/:examId/join` (login + masukkan token, membuat/ambil `exam_session`)
- `GET  /api/v1/student/sessions/:sessionId` (state + timer)
- `GET  /api/v1/student/sessions/:sessionId/questions` (soal yang sudah di-assemble; tidak mengirim kunci jawaban)
- `GET  /api/v1/student/sessions/:sessionId/answers` (ambil jawaban tersimpan untuk resume)
- `POST /api/v1/student/sessions/:sessionId/answers` (upsert jawaban)
- `POST /api/v1/student/sessions/:sessionId/submit`
- `POST /api/v1/student/sessions/:sessionId/heartbeat` (opsional payload JSON)
- `GET  /api/v1/student/results` (riwayat hasil + auto-score)
- `GET  /api/v1/student/announcements` (pengumuman aktif untuk siswa; filter: `q`, `limit`, `offset`)
- `POST /api/v1/student/attendance` (submit/update absensi ujian)
- `GET  /api/v1/student/attendance/history` (riwayat absensi; filter: `q`, `limit`, `offset`)

Contoh join:

```json
{"token":"ABC123"}
```

Contoh simpan jawaban:

```json
{
  "question_id": "uuid",
  "answer_json": {"selected":["A"]}
}
```
