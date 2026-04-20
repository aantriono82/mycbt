# Blueprint Implementasi mycbt (Gin + Vue Admin One)

Dokumen ini adalah panduan implementasi end-to-end untuk membangun aplikasi CBT dari scaffold repo ini. Targetnya: cepat jadi, mudah dirawat, dan aman untuk produksi.

## 1) Scope & Peran Pengguna

### Peran
- Admin: mengelola master data, bank soal global, jadwal ujian, token, monitoring, evaluasi, cetak, settings.
- Guru: mengelola bank soal mapel yang diampu, jadwal ujian kelas yang diajar, token, monitoring, evaluasi, profil.
- Siswa: melihat jadwal ujian, masuk ruang ujian, mengerjakan ujian (timer + navigasi), melihat hasil, profil, pengumuman/absensi (opsional).

### Menu (acuan)
Frontend sudah disiapkan routing/menu per role:
- Admin: `/admin/*`
- Guru: `/teacher/*`
- Siswa: `/student/*`

Catatan: saat ini role masih “mock” (dipilih dari navbar) untuk mempercepat pengembangan UI. Nantinya diganti oleh autentikasi yang benar.

## 2) Arsitektur Target

### Komponen
- Backend API (Gin):
  - REST API v1: `/api/v1/*`
  - Auth (JWT atau session), RBAC, audit log
  - Upload & import file: Excel untuk master data, DOCX untuk bank soal
  - Realtime monitoring ujian: SSE atau WebSocket
- Frontend (Vue 3 + Pinia + Vue Router):
  - Layout admin dashboard + role-based navigation
  - Modul Master Data, Bank Soal, Ujian, Monitoring, Evaluasi, Cetak, Settings

### Komunikasi
- HTTP JSON untuk CRUD & workflow.
- Realtime:
  - Opsi 1: SSE (Server-Sent Events) untuk monitor (lebih sederhana, cukup 1 arah).
  - Opsi 2: WebSocket untuk kebutuhan 2 arah (mis. broadcast token, force submit).

Rekomendasi awal: SSE untuk monitor; WebSocket ditambah kalau nanti ada fitur “push command”.

## 3) Struktur Repo & Konvensi

### Backend (`backend/`)
- `cmd/api`: entrypoint server.
- `internal/config`: config env.
- `internal/httpapi`: router/handler.

Konvensi yang disarankan saat bertambah:
- `internal/domain`: entity + rules (pure business).
- `internal/repo`: akses DB (sqlc/gorm).
- `internal/service`: use-case (orchestrator).
- `internal/httpapi/handlers`: grouping per resource.
- `internal/httpapi/middleware`: auth, rbac, request-id, rate limit.

### Frontend (`frontend/`)
- `src/menuAside.js`: generator menu per role.
- `src/menuNavBar.js`: menu navbar (termasuk switch role sementara).
- `src/router/index.js`: rute per role.
- `src/services/api.js`: Axios client (`VITE_API_BASE_URL`).
- `src/stores/auth.js`: role store (sementara).

Konvensi yang disarankan:
- `src/modules/<module>`: setiap modul punya `pages/`, `components/`, `api.js`.
- `src/stores/`: pinia stores per domain (users, exams, questionBank, monitoring).
- `src/views/`: role dashboards dan wrapper views.

## 4) Data Model (Rancangan Database)

Rekomendasi DB: PostgreSQL.

### Master Data
- `schools`
  - `id`, `name`, `logo_url`, `address`, `created_at`, `updated_at`
- `users`
  - `id`, `username`, `password_hash`, `role` (admin/teacher/student), `is_active`
  - `name`, `email`, `phone`
  - `school_id`, `created_at`, `updated_at`
- `teachers`
  - `id`, `user_id`, `nip` (optional), `school_id`
- `students`
  - `id`, `user_id`, `nis` (unique), `school_id`
  - `program_id`, `level_id`, `group_id`
- `programs`
  - `id`, `school_id`, `code`, `name`
- `levels`
  - `id`, `school_id`, `name` (mis. Kelas 10/11/12)
- `groups`
  - `id`, `school_id`, `name` (mis. X IPA-1)
- `subjects`
  - `id`, `school_id`, `code`, `name`
- `teacher_subjects`
  - `teacher_id`, `subject_id` (mapping mapel yang diampu)

### Bank Soal
- `question_sets` (paket/bank soal)
  - `id`, `school_id`, `subject_id`, `owner_teacher_id` (nullable untuk global/admin)
  - `title`, `status` (draft/published), `created_at`
- `questions`
  - `id`, `question_set_id`, `type`, `stem` (HTML/markdown), `difficulty`, `order_no`
  - `attachment_url` (optional)
- `question_options` (opsi)
  - `id`, `question_id`, `label` (A/B/C), `content`, `is_correct`
- `question_matches` (menjodohkan)
  - `id`, `question_id`, `left_content`, `right_content`, `right_key`
- `question_answers` (kunci untuk isian singkat / benar-salah)
  - `id`, `question_id`, `answer_text`, `answer_bool`
- `question_essay_rubrics` (opsional)
  - `id`, `question_id`, `rubric_text`, `max_score`

Catatan tipe soal (6 tipe):
1. Pilihan Ganda Single: 1 opsi benar.
2. Pilihan Ganda Multiple: banyak opsi benar.
3. Menjodohkan: pasangan left->right.
4. Isian Singkat: string matching (perlu normalisasi).
5. Uraian: manual grading (atau rubric).
6. Benar Salah: boolean.

### Ujian
- `exams`
  - `id`, `school_id`, `title`, `subject_id`, `teacher_id`
  - `starts_at`, `ends_at`, `duration_minutes`
  - `shuffle_questions`, `shuffle_options`
  - `status` (draft/published/archived)
- `exam_question_sets` (jika ujian mengambil dari 1+ bank)
  - `exam_id`, `question_set_id`, `num_questions` (optional)
- `exam_targets`
  - `exam_id`, `level_id` (nullable), `group_id` (nullable), atau `student_id` (nullable)
  - (pilih salah satu strategi targeting; pastikan constraint)
- `exam_tokens`
  - `id`, `exam_id`, `token`, `valid_from`, `valid_to`, `is_active`, `created_by`
- `exam_sessions`
  - `id`, `exam_id`, `student_id`
  - `started_at`, `finished_at`, `status` (in_progress/submitted/forced/expired)
  - `client_ip`, `user_agent`, `device_fingerprint` (optional)
- `exam_attempts` (jawaban per soal)
  - `id`, `exam_session_id`, `question_id`
  - `answer_json` (pilihan, pasangan, teks, bool)
  - `score_auto`, `score_manual`, `answered_at`
- `exam_events` (monitoring/audit)
  - `id`, `exam_session_id`, `type` (login/start/answer/heartbeat/submit/disconnect)
  - `payload_json`, `created_at`

### Evaluasi & Cetak
- `exam_results_view` (bisa berupa materialized view)
  - agregasi nilai per siswa/per ujian
- `reports_exports`
  - tracking export excel (siapa, kapan, parameter)

## 5) Backend API (Rancangan Endpoint)

Konvensi: semua API di `/api/v1`, JSON response standar:
- `{"data":..., "meta":...}` untuk sukses.
- `{"error": {"code": "...", "message": "...", "details": ...}}` untuk error.

### Auth
- `POST /api/v1/auth/login` -> JWT access + refresh (atau session cookie)
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET  /api/v1/me` -> profil + role + permissions

### Master Data (Admin)
- Guru:
  - `GET /api/v1/admin/teachers`
  - `POST /api/v1/admin/teachers`
  - `POST /api/v1/admin/teachers/import` (Excel)
  - `GET /api/v1/admin/teachers/template` (download template)
- Siswa:
  - `GET /api/v1/admin/students`
  - `POST /api/v1/admin/students`
  - `POST /api/v1/admin/students/import` (Excel)
  - `GET /api/v1/admin/students/template`
- Program/Level/Group/Mata Pelajaran: CRUD standar.

### Bank Soal (Admin/Guru)
- `GET /api/v1/question-sets` (filter by subject, owner)
- `POST /api/v1/question-sets`
- `GET /api/v1/question-sets/:id`
- `POST /api/v1/question-sets/:id/questions`
- `POST /api/v1/question-sets/:id/import-docx/preview` (upload `.docx` -> preview)
- `POST /api/v1/question-sets/:id/import-docx` (upload `.docx` -> insert)
- `PATCH /api/v1/questions/:id` (update)
- `DELETE /api/v1/questions/:id`

### Ujian (Admin/Guru)
- Jadwal:
  - `GET /api/v1/exams`
  - `POST /api/v1/exams`
  - `GET /api/v1/exams/:id`
  - `PATCH /api/v1/exams/:id`
  - `DELETE /api/v1/exams/:id`
  - `GET/PUT /api/v1/exams/:id/targets`
  - `GET/PUT /api/v1/exams/:id/question-sets`
- Token:
  - `POST /api/v1/exams/:id/tokens` (generate)
  - `GET /api/v1/exams/:id/tokens`
  - `PATCH /api/v1/tokens/:tokenId` (enable/disable)
- Monitor:
  - `GET /api/v1/exams/:id/monitor/sessions` (snapshot)
  - `GET /api/v1/exams/:id/monitor/stream` (SSE)
- Reset Login:
  - `POST /api/v1/exams/:id/sessions/:sessionId/reset` (optional)

### Siswa (Ruang Ujian)
- `GET /api/v1/student/exams` (ujian yang available sesuai jadwal & target)
- `POST /api/v1/student/exams/:examId/join` (token + verifikasi jadwal)
- `GET /api/v1/student/sessions/:sessionId` (state, timer)
- `GET /api/v1/student/sessions/:sessionId/questions` (soal yang sudah di-assemble, aman)
- `POST /api/v1/student/sessions/:sessionId/answers` (save answer)
- `POST /api/v1/student/sessions/:sessionId/submit`
- `POST /api/v1/student/sessions/:sessionId/heartbeat` (monitor online)

### Evaluasi
- `GET /api/v1/exams/:id/results` (rekap)
- `GET /api/v1/exams/:id/item-analysis` (analisis butir)
- `POST /api/v1/exams/:id/export` (excel)

### Settings
- `GET /api/v1/settings/school-identity`
- `PUT /api/v1/settings/school-identity`
- `GET /api/v1/settings/system`
- `PUT /api/v1/settings/system`

## 6) Import Excel (Guru/Siswa)

### Template
Sediakan endpoint `.../template` yang mengembalikan file `.xlsx` template:
- Sheet `README`: aturan pengisian, contoh.
- Sheet `DATA`: header baku.

Rekomendasi format:
- Guru: `nip`, `nama`, `email`, `username` (optional), `mapel_codes` (comma separated).
- Siswa: `nis`, `nama`, `email`, `program_code`, `level`, `group`, `username` (optional).

### Validasi
- Validasi header wajib.
- Validasi unique key: `nis`, `nip`, `username`.
- Preview import (opsional):
  - `POST /import?dry_run=1` -> kembalikan daftar row valid/invalid.

### Library
Go: `excelize` atau `tealeg/xlsx`. Pilih `excelize` untuk fitur luas.

## 7) Upload Soal DOCX + Parser

### UX
Halaman Bank Soal:
- Pilih Mapel + Bank Soal
- Upload `.docx`
- Preview hasil parsing: daftar soal + tipe + kunci
- Konfirmasi import

### Desain format dokumen (rekomendasi)
Karena docx bebas format, harus ada aturan yang konsisten. Contoh:
- Soal dimulai dengan `1.` atau `1)`
- Pilihan:
  - `A. ...`, `B. ...`
  - Kunci: `KUNCI: A` atau `KUNCI: A,C` (untuk multiple)
- Benar/Salah:
  - `JAWABAN: BENAR` / `JAWABAN: SALAH`
- Menjodohkan:
  - `PASANGAN:` lalu daftar `1) ... = a) ...`
- Isian singkat:
  - `JAWABAN: ...` (bisa beberapa variasi dipisah `|`)

### Implementasi parser
Go: gunakan library `baliance/gooxml` (DOCX) atau konversi ke plain text + heuristik.
Tahap awal:
1. Extract text + run-level formatting minimal.
2. Parse blok soal berdasarkan nomor.
3. Deteksi tipe berdasarkan pattern opsi/pasangan/jawaban.
4. Simpan hasil sebagai draft; user review sebelum publish.

## 8) Engine Ujian (Timer, Shuffle, Keamanan)

### Assembling soal
Ketika siswa `join`:
- Validate jadwal (waktu server).
- Validate token (jika diperlukan).
- Buat `exam_session`.
- Assemble daftar soal:
  - Ambil dari `question_sets` sesuai rule.
  - Shuffle (seed = session id) agar deterministik.
  - Simpan mapping `session_question_order` (bisa tabel baru atau payload json).

### Timer
Sumber kebenaran timer: server.
- Server menyimpan `started_at`.
- Client menampilkan countdown, tapi submit/expired ditentukan server.

### Save jawaban
Gunakan upsert per question dalam session. Kirim minimal payload:
- `question_id`, `answer_json`, `client_ts` (optional).

### Submit
Saat submit:
- Lock session: set `finished_at`, status `submitted`.
- Auto grading untuk tipe objective:
  - single, multiple, matching, true/false, short answer (dengan normalisasi)
- Essay: `score_manual` diisi guru.

### Keamanan dasar (minimal)
- Rate limit login/join.
- Restrict `GET questions` hanya untuk session aktif milik siswa.
- Cegah melihat kunci di payload.
- Audit event untuk aktivitas penting.

## 9) Monitoring Real-time

### Data yang ditampilkan
Monitor Ujian:
- jumlah online, in_progress, submitted, disconnected
- timeline aktivitas

Monitor Peserta:
- list peserta: status online, last_seen, progress (soal dijawab/total), IP/device (opsional)

### Mekanisme
SSE endpoint:
- `GET /api/v1/exams/:id/monitor/stream`
- Event type:
  - `session_started`, `heartbeat`, `answer_saved`, `session_submitted`, `session_disconnected`

Client:
- buka 1 koneksi SSE saat halaman monitor aktif.

## 10) Evaluasi / Analisis Butir Soal

Output minimal:
- Nilai per siswa
- Distribusi nilai
- Item analysis:
  - tingkat kesukaran (p-value)
  - daya pembeda (mis. korelasi point-biserial)
  - distraktor (pilihan mana yang dipilih)

Catatan: untuk soal essay, item analysis parsial (butuh skor manual).

## 11) Cetak (Kartu Ujian, Daftar Hadir, Laporan Nilai)

Rekomendasi:
- Generate HTML server-side atau frontend template, lalu export PDF:
  - Opsi 1: backend render HTML + wkhtmltopdf
  - Opsi 2: frontend print stylesheet + “Print to PDF”
Tahap awal: implement “print-friendly page” di frontend.

## 12) Settings

Identitas sekolah:
- logo upload (storage lokal/S3-compatible)
- nama, alamat, kontak

Pengaturan sistem (contoh):
- timezone sekolah
- mode token required
- kebijakan reset login
- batas device / sesi

## 13) Milestone Implementasi (Rekomendasi)

### M0 (sudah)
- Scaffold backend + frontend template, role-based menu, routing, health check.

### M1: Auth + RBAC (wajib sebelum produksi)
- Login (admin/guru/siswa), JWT, `/me`
- Route guard di frontend (role-based)
- Logout

### M2: Master Data (Admin)
- CRUD Program/Level/Group/Mapel
- Import Guru + Import Siswa (Excel) + template download

Status repo saat ini:
- Endpoint Master Data dasar sudah tersedia di `/api/v1/admin/*` (admin-only, JWT).
- Template + import Excel Guru/Siswa sudah tersedia.

### M3: Bank Soal (Guru/Admin)
- CRUD bank soal + CRUD soal manual
- Upload DOCX + preview + import

### M4: Ujian
- CRUD jadwal ujian + targeting (group/level/student)
- Token generate
- Siswa: join, lihat soal, save answer, submit, timer

### M5: Monitoring
- Heartbeat siswa
- SSE monitor
- Reset login

### M6: Evaluasi + Export
- Rekap nilai
- Export Excel
- Item analysis basic

### M7: Cetak
- Kartu ujian, daftar hadir, laporan nilai

## 14) Checklist Kualitas

Backend:
- request id, structured logging
- migrations (golang-migrate)
- OpenAPI spec (opsional tapi sangat membantu)
- unit test untuk scoring & parser

Frontend:
- route guard + fallback UX
- form validation
- loading/error states untuk semua halaman

Security:
- password hashing (bcrypt/argon2)
- CSRF jika pakai cookies
- input validation & upload size limit
- audit log untuk tindakan admin/guru
