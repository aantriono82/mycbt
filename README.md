# AtigaCBT

[![Backend Tests](https://github.com/aantriono82/mycbt/actions/workflows/backend-tests.yml/badge.svg)](https://github.com/aantriono82/mycbt/actions/workflows/backend-tests.yml)

Scaffold aplikasi CBT (Computer Based Test) dengan:

- Backend: Gin (Go) in `backend/`
- Frontend: Admin One Vue Tailwind in `frontend/`

Dokumen implementasi lengkap: lihat [docs/BLUEPRINT.md](/home/aantriono/dev/atigacbt/docs/BLUEPRINT.md).
Kontrak API awal (OpenAPI): [docs/openapi.yaml](/home/aantriono/dev/atigacbt/docs/openapi.yaml).
Panduan deploy production: [docs/DEPLOYMENT.md](/home/aantriono/dev/atigacbt/docs/DEPLOYMENT.md).
Checklist predeploy: [docs/PREDEPLOY_CHECKLIST.md](/home/aantriono/dev/atigacbt/docs/PREDEPLOY_CHECKLIST.md).

## Status Implementasi (2026-04-23)

- **Modernisasi UI/UX (Skeleton Screens)**: Implementasi komponen `BaseSkeleton.vue` pada Dashboard Siswa, Pusat Pengumuman, dan Evaluasi Admin untuk menghilangkan *layout shift* dan memberikan feedback visual yang premium saat data sedang dimuat.
- **Optimasi Data dengan TanStack Query**: Integrasi `@tanstack/vue-query` untuk manajemen *server state*. Memungkinkan percepatan akses data melalui *automatic caching*, sinkronisasi latar belakang, dan pengurangan beban request ke server.
- **Real-time Notifications (SSE)**: Implementasi *streaming* notifikasi via *Server-Sent Events* (SSE) di backend dan frontend. Siswa kini mendapatkan update instan untuk pengumuman atau perubahan jadwal ujian tanpa perlu memuat ulang halaman.
- **Progressive Web App (PWA) & Fullscreen Mode**: Aplikasi kini dapat diinstal sebagai PWA di desktop/mobile. Menyertakan fitur *Fullscreen Mode* pada lembar kerja ujian untuk membantu siswa fokus dan meminimalisir pembukaan tab lain selama durasi ujian.
- **Centralized Error Handling & Global Toast**: Implementasi Axios interceptor untuk menangani error secara global (401, 403, 500, network error). Dilengkapi dengan sistem notifikasi *toast* (Pinia-based) yang seragam di seluruh aplikasi, menggantikan blok try-catch yang redundan.
- **Pinia State Refactoring**: Sentralisasi logika operasional ujian (`examStore`) dan hasil ujian (`resultStore`) ke Pinia, memisahkan logika bisnis dari komponen UI agar kode lebih ringkas dan mudah dipelihara.
- **Stabilisasi & Linting**: Perbaikan bug deklarasi ganda pada layout dan sinkronisasi status token ujian.
- **Bulk Approve Verifikasi Pendaftaran**: Halaman `/admin/master-data/verifikasi-pendaftaran` kini memiliki aksi `Approve Semua Sesuai Filter` untuk memproses pendaftaran `pending` secara massal.
- **Paritas Panel Admin/Guru**: Route admin dan guru memakai view operasional yang sama dengan RBAC yang ketat.
- **Timezone-Aware Session Management**: Validasi jendela waktu pengerjaan akurat mengikuti zona waktu lokal (WIB/WITA/WIT).

Backend yang sudah berjalan:
- Auth JWT (`/api/v1/auth/login`, `/api/v1/me`) + RBAC sederhana (admin/teacher/student).
- Request ID (`X-Request-ID`) otomatis untuk seluruh response API.
- Structured logging JSON per request (termasuk `request_id`, status, latency, user role/id jika tersedia).
- Audit log tindakan mutasi (`POST/PUT/PATCH/DELETE`) role admin/guru ke tabel `audit_logs`.
- Unit test untuk parser soal DOCX/plain-text, rate limit middleware, dan evaluator scoring engine per tipe soal sudah tersedia.
- Rate limit dasar untuk endpoint sensitif.
- Master Data (admin): CRUD lengkap + import Excel.
- Bank Soal: CRUD + parser DOCX + preview.
- Ujian: Manajemen jadwal, token, dan monitoring.

Frontend yang sudah berjalan:
- Login & Sinkronisasi Role via JWT.
- Dashboard Admin, Guru, dan Siswa dengan statistik riil dari backend.
- UI/UX Premium: Menggunakan desain "Atiga Premium" dengan kontras tinggi dan mode gelap yang dioptimalkan.
- Bank Soal: Quick Add Soal (Batch) lewat kartu "Tambah Soal" dan tombol header "Tambah Banyak" di editor.
- Modul lengkap Admin: Master Data, Verifikasi Pendaftaran, Bank Soal, Jadwal, Token, Monitor, Absensi, Cetak, dan Settings.
- Modul lengkap Siswa: Ruang Ujian, Kerjakan Ujian (autosave), Hasil Ujian, dan Pengumuman.
- Integrasi backend untuk seluruh modul utama.
## Prereqs

- Docker Engine + Docker Compose Plugin
- Node.js `^20.19.0` atau `>=22.12.0` (cek: `node -v`)
- npm (ikut dari Node.js)
- Go: install lokal (tanpa sudo) via:

```bash
./scripts/install-go.sh
```

### Setup cepat (Linux Mint/Ubuntu fresh install)

Install dependency sistem:

```bash
sudo apt update
sudo apt install -y docker.io docker-compose-v2 curl git
```

Aktifkan Docker:

```bash
sudo systemctl enable --now docker
sudo usermod -aG docker $USER
```

Lalu logout/login sekali agar group `docker` aktif.

Jika `node -v` masih 18.x, upgrade ke Node 22:

```bash
curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
sudo apt install -y nodejs
```

## Run (dev)

Cara termudah untuk menjalankan aplikasi (DB, Backend, dan Frontend):

```bash
./run-local.sh
```

## Testing Backend

Untuk menjalankan suite backend dari root repo:

```bash
make test-backend
```

Integration test repo PostgreSQL bisa dijalankan dengan:

```bash
export TEST_DATABASE_URL='postgres://user:pass@localhost:5432/dbname?sslmode=disable'
make test-backend-integration
```

Dokumentasi detail ada di [backend/README.md](/home/aantriono/dev/atigacbt/backend/README.md).

## Optimasi Ukuran Workspace dan VPS

Agar repo dan server tetap ringan:

1. Bersihkan cache/build lokal secara berkala:
```bash
./scripts/cleanup-workspace.sh
```
2. Jangan commit binary hasil build (`api_bin`, `backend/api*`, `seed`, `migrate`) atau cache lokal.
3. Untuk production gunakan image dari `deploy/compose.production.yml` (multi-stage Dockerfile sudah dioptimasi dengan Go stripped binary).

## Branch Protection

Rekomendasi minimum untuk branch `main`:
- require pull request sebelum merge
- block force push dan branch deletion
- require status checks:
  - `Unit and Handler Tests`
  - `PostgreSQL Integration Tests`
- require branch up to date sebelum merge jika tim sering merge paralel
- require conversation resolution sebelum merge

Kalau repo ini dipakai tim kecil dan ingin tetap cepat, saya tidak sarankan dulu mewajibkan banyak reviewer. Status check backend jauh lebih penting daripada approval count tinggi.

Catatan (first time / DB baru): `run-local.sh` hanya menyalakan proses. Jika tabel belum ada atau user admin belum dibuat, jalankan migrate + seed sekali:

```bash
cd backend
export DATABASE_URL="postgres://atigacbt:atigacbt@localhost:5433/atigacbt?sslmode=disable"
export JWT_SECRET="7f59f6b9c9f2b8e8a8b8c8d8e8f808182838485868788898a8b8c8d8e8f8081"

../.tooling/go/bin/go run ./cmd/migrate
export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=admin12345
../.tooling/go/bin/go run ./cmd/seed
```

### Manual Run (Advanced)

Jika ingin menjalankan secara terpisah:

#### 1. Start Postgres
```bash
docker compose up -d
```

#### 2. Terminal 1: Backend (migrate + seed + run API)
```bash
cd backend
export DATABASE_URL="postgres://atigacbt:atigacbt@localhost:5433/atigacbt?sslmode=disable"
export JWT_SECRET="7f59f6b9c9f2b8e8a8b8c8d8e8f808182838485868788898a8b8c8d8e8f8081"
../.tooling/go/bin/go run ./cmd/migrate

export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=admin12345
../.tooling/go/bin/go run ./cmd/seed

../.tooling/go/bin/go run ./cmd/api
```

#### 3. Terminal 2: Frontend
```bash
cd frontend
# optional: cp .env.example .env lalu set VITE_AP --portI_BASE_URL
npm run dev -- --port 5173
```

Catatan:
- Jika muncul 2 URL frontend (`5173` dan `5174`), artinya ada lebih dari satu proses Vite yang berjalan.
- Matikan proses lama sebelum start ulang frontend:
```bash
pkill -f vite
```

## Audit Panel Admin (Backend)

Untuk memastikan endpoint admin tetap konsisten dengan UI (regresi cepat ketahuan), jalankan smoke audit berikut (pastikan API sudah running):

```bash
BASE_URL=http://127.0.0.1:8080 ADMIN_USERNAME=admin ADMIN_PASSWORD=admin12345 bash scripts/audit_admin.sh
```

Indikator sukses: output berakhir dengan `ALL OK`.


Default URL:

- Frontend: `http://localhost:5173/admin-one-vue-tailwind/`
- Backend: `http://localhost:8080`
- Health: `GET http://localhost:8080/healthz`
- Ping: `GET http://localhost:8080/api/v1/ping`

Login (UI):
- Buka `http://localhost:5173/admin-one-vue-tailwind/`
- Default admin (setelah seed): username `admin`, password `admin12345`

Auth (JWT):
- `POST http://localhost:8080/api/v1/auth/login`
- `GET  http://localhost:8080/api/v1/me` (requires `Authorization: Bearer <token>`)

Contoh alur siswa: login -> lihat ujian -> join pakai token ujian:

```bash
# 1) Login (ambil JWT access_token)
curl -sS -X POST http://localhost:8080/api/v1/auth/login \
  -H 'content-type: application/json' \
  -d '{"username":"siswa1","password":"password-siswa"}'

# 2) Pakai token untuk request berikutnya
# (opsional pakai jq) TOKEN="$(curl ... | jq -r '.data.access_token')"
export TOKEN="paste_access_token_di_sini"

# 3) List ujian yang ditargetkan ke siswa (lihat `id`)
curl -sS http://localhost:8080/api/v1/student/exams \
  -H "Authorization: Bearer $TOKEN"

# 4) Join ujian dengan token ujian (token dibuat admin/guru dari menu Token)
export EXAM_ID="paste_exam_id_di_sini"
curl -sS -X POST "http://localhost:8080/api/v1/student/exams/$EXAM_ID/join" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'content-type: application/json' \
  -d '{"token":"123456"}'

# Response berisi `data.session.id` (SESSION_ID). Setelah sudah join dan status masih `in_progress`,
# panggilan join berikutnya tidak perlu input token lagi.
```

Master Data (admin-only; requires JWT):
- CRUD: `programs`, `levels`, `groups`, `subjects`
- CRUD: `announcements` (pengumuman siswa dengan target opsional level/group/student)
- CRUD: `teachers`, `students` (manual create + import Excel template)
- CRUD: `sessions` (master sesi dengan jam mulai/selesai)
- Mapel guru: `GET/PUT /api/v1/admin/teachers/:id/subjects`

Registrasi & Verifikasi (Pilihan B: Manual -> Approve -> Google Link):
- **Alur Pendaftaran**: Calon pengguna mengisi form manual 3-langkah (tanpa login Google di awal).
- **Public API**: `POST /api/v1/auth/google/register` (tanpa mewajibkan `google_id`).
- **Verifikasi Admin**: Admin meninjau data lengkap (NISN, No HP, Sekolah) di panel `/admin/master-data/verifikasi-pendaftaran`.
- **Bulk Approve**: Admin bisa approve massal dengan filter aktif via `POST /api/v1/admin/registrations/approve-bulk` (opsional payload: `role`, `q`, `limit`, `note`).
- **Auto-Linking**: Setelah disetujui, pendaftar login menggunakan tombol Google. Sistem akan mencocokkan akun via **Email**, menautkan `google_id` secara otomatis, dan memberikan akses masuk.
- **Set Pending**: Fitur baru bagi admin untuk mengembalikan status `approved/rejected` menjadi `pending` demi fleksibilitas data.

Bank Soal (tahap awal; admin/teacher; requires JWT):
- `GET/POST /api/v1/question-sets`
- `GET/PATCH/DELETE /api/v1/question-sets/:id`
- `GET/POST /api/v1/question-sets/:id/questions`
- `GET /api/v1/questions/:id` (ambil 1 soal)
- `PATCH /api/v1/questions/:id` (edit soal + payload per tipe)
- `DELETE /api/v1/questions/:id`
- `POST /api/v1/question-sets/:id/import-docx/preview` (upload `.docx` -> preview parse)
- `POST /api/v1/question-sets/:id/import-docx` (upload `.docx` -> insert questions)

Ujian (jadwal + token; admin/teacher; requires JWT):
- `GET/POST /api/v1/exams`
- `GET/PATCH/DELETE /api/v1/exams/:id`
- `GET/PUT /api/v1/exams/:id/targets` (target level/group/siswa)
- `GET/PUT /api/v1/exams/:id/question-sets` (attach bank soal)
- `GET/POST /api/v1/exams/:id/tokens`
- `PATCH /api/v1/tokens/:id` (enable/disable token)
- `GET /api/v1/exams/:id/results` (rekap peserta + auto-score; teacher: hanya ujian miliknya)
- `GET /api/v1/exams/:id/item-analysis` (analisis butir: p-value, tingkat kesukaran, daya pembeda upper/lower group, distraktor opsi)
- `GET /api/v1/exams/:id/score-distribution` (distribusi nilai + min/avg/median/max)
- `GET /api/v1/exams/:id/export` (export `.xlsx` multi-sheet: ExecutiveSummary + Results + Score Distribution + Item Analysis)
- `GET /api/v1/exams/:id/item-analysis/export` (export analisis butir + d-index + distraktor ke `.xlsx`)

## Sistem Pembobotan Soal dan Konversi Nilai

### Ringkasannya
- Bobot disimpan per soal pada field `weight` (minimal `> 0`).
- Nilai akhir peserta selalu dinormalisasi ke skala `0-100`.
- Mode penilaian ujian yang direkomendasikan untuk skema HOTS: `partial`.

### Preset Bobot HOTS/Analitis (per tipe soal)
- `mc_single` (Pilihan Ganda): `1`
- `mc_multiple` (PG Kompleks): `2`
- `matching` (Menjodohkan): `2`
- `short_answer` (Isian Singkat): `1`
- `true_false` (Benar/Salah): `1`
- `essay` (Esai/Uraian): `3`

### Rumus Konversi ke Nilai Akhir
Setiap soal dinilai dulu dalam rentang `0-100`, lalu dihitung berbobot:

`Nilai Akhir = round((sum((skor_soal/100) * bobot_soal) / sum(bobot_soal)) * 100)`

Bentuk ekuivalen:

`Nilai Akhir = round(sum(skor_soal * bobot_soal) / sum(bobot_soal))`

Catatan:
- `sum(skor_soal * bobot_soal)` adalah total poin berbobot (contoh angka seperti `3320` berasal dari sini).
- `round(...)` mengikuti pembulatan ke bilangan bulat terdekat.

### Contoh Kasus (30 Soal Campuran)
Komposisi soal:
- 10 PG, 6 PG Kompleks, 4 Menjodohkan, 5 Isian, 3 Benar/Salah, 2 Esai.

Dengan preset HOTS, total bobot maksimum:
- `10*1 + 6*2 + 4*2 + 5*1 + 3*1 + 2*3 = 44`

Jika jawaban benar:
- 5 PG, 3 PG Kompleks, 2 Menjodohkan, 2 Isian, 1 Benar/Salah, 1 Esai.

Maka bobot tercapai:
- `5*1 + 3*2 + 2*2 + 2*1 + 1*1 + 1*3 = 21`

Hasil:
- Skor berbobot: `21 dari 44` (setara `2100 dari 4400` pada skala poin berbobot).
- Nilai akhir: `(21/44)*100 = 47.73`, dibulatkan menjadi `48`.
- Supaya konsisten HOTS, set juga **Mode Penilaian ujian = `partial`** (agar PG kompleks/menjodohkan/benar-salah bisa dapat nilai proporsional, tidak 0/100 penuh).
- `GET /api/v1/exams/:id/attendance` (rekap absensi peserta + persentase kehadiran; teacher: hanya ujian miliknya)
- `GET /api/v1/exams/:id/monitor/sessions` (snapshot sesi peserta + progress; polling-friendly)
- `GET /api/v1/exams/:id/monitor/participants` (snapshot peserta target + status join/online/progress; polling-friendly)
- `GET /api/v1/exams/:id/monitor/stream?view=sessions|participants&access_token=...` (SSE stream; untuk EventSource karena tidak bisa set header Authorization)
- `POST /api/v1/exams/:id/sessions/:sessionId/reset` (reset login: hapus session non-submitted agar siswa bisa join ulang)
- `POST /api/v1/exams/:id/sessions/:sessionId/force-submit` (paksa submit sesi `in_progress`; status akhir `forced`)
- `GET /api/v1/exams/:id/sessions/:sessionId/essays` (ambil list jawaban essay untuk dikoreksi)
- `POST /api/v1/exams/:id/sessions/:sessionId/essays/score` (simpan nilai & feedback essay)

Announcements (admin/teacher):
- `GET/POST /api/v1/announcements`
- `GET/PATCH/DELETE /api/v1/announcements/:id`
- `GET /api/v1/lookups/levels`, `groups`, `students` (untuk dropdown target ujian/pengumuman)
  - Role `teacher`: otomatis terfilter berdasarkan penugasan (teacher assignments).

Maintenance (admin-only):
- `GET /api/v1/maintenance/backup` (ekspor database ke .sql)
- `POST /api/v1/maintenance/restore` (impor database dari .sql)
- `GET /api/v1/lms/export/students` (ekspor CSV/JSON roster siswa)
- `GET /api/v1/lms/export/results` (ekspor CSV/JSON hasil ujian)
- `GET /api/v1/analytics/dashboard` (data tren dan komparasi analitik)
- `GET/POST /api/v1/lti/platforms` (manajemen platform LMS)
- `GET /api/v1/lti/login` & `POST /api/v1/lti/launch` (public OIDC/LTI endpoints)

Settings (admin; requires JWT):
- `GET /api/v1/settings/school-identity`
- `PUT /api/v1/settings/school-identity`
- `GET /api/v1/settings/system`
- `PUT /api/v1/settings/system`
  - `token_required` kini diterapkan pada alur `POST /api/v1/student/exams/:examId/join`:
    - `true`: token wajib (perilaku default)
    - `false`: token opsional (jika dikirim tetap diverifikasi)
  - `max_active_sessions` kini diterapkan pada alur `POST /api/v1/student/exams/:examId/join`:
    - jika sesi `in_progress` siswa sudah mencapai limit, join baru ditolak `409 max_active_sessions_reached`
  - `allow_reset_login` kini diterapkan pada `POST /api/v1/exams/:id/sessions/:sessionId/reset`:
    - `true`: reset login diizinkan
    - `false`: endpoint reset akan menolak dengan `403 policy_disabled`
  - `attendance_require_ip` kini diterapkan pada `POST /api/v1/student/attendance`:
    - `true`: request wajib memiliki `client_ip` valid (dideteksi server)
    - data absensi kini menyimpan `client_ip` (nullable)

Catatan: siswa hanya dapat melihat ujian dengan status `published`.

Siswa (ruang ujian; requires JWT role `student`):
- `GET  /api/v1/student/exams` (ujian yang ditargetkan ke siswa; pagination + `q`)
- `GET  /api/v1/student/exams/:examId/session` (cek sesi aktif untuk resume tanpa input token ulang)
- `POST /api/v1/student/exams/:examId/join` (login + input token -> buat/ambil `exam_session`)
- `GET  /api/v1/student/sessions/:sessionId` (state + timer)
- `GET  /api/v1/student/sessions/:sessionId/questions` (soal yang sudah di-assemble; tanpa kunci jawaban)
- `GET  /api/v1/student/sessions/:sessionId/answers` (load jawaban tersimpan untuk resume)
- `POST /api/v1/student/sessions/:sessionId/answers` (upsert jawaban)
- `POST /api/v1/student/sessions/:sessionId/submit`
- `POST /api/v1/student/sessions/:sessionId/heartbeat` (opsional payload JSON)
- `GET  /api/v1/student/results` (riwayat sesi submitted/expired + auto-score)
- `GET  /api/v1/student/announcements` (pengumuman aktif yang relevan untuk siswa; pagination + `q`)
- `POST /api/v1/student/attendance` (submit/update absensi ujian)
- `GET  /api/v1/student/attendance/history` (riwayat absensi; pagination + `q`)

## Struktur Proyek

- `backend/`: service API Gin.
- `frontend/`: Vue 3 + Vite + Tailwind (template `admin-one-vue-tailwind`).
- `scripts/install-go.sh`: installer Go lokal ke `.tooling/go`.
- Dokumen tambahan:
  - Backend: [backend/README.md](/home/aantriono/dev/atigacbt/backend/README.md)
  - Frontend: [frontend/README.md](/home/aantriono/dev/atigacbt/frontend/README.md)

## Menu & Role (Frontend)

Frontend sudah disiapkan routing terpisah per role:

- Admin: prefix `/admin/*`
- Guru: prefix `/teacher/*`
- Siswa: prefix `/student/*`

Menu sidebar akan berubah sesuai role (sementara role bisa diganti dari dropdown di navbar dan tersimpan di `localStorage`).

Shortcut rute:
- Admin dashboard: `/admin-one-vue-tailwind/#/admin/dashboard`
- Guru dashboard: `/admin-one-vue-tailwind/#/teacher/dashboard`
- Siswa dashboard: `/admin-one-vue-tailwind/#/student/dashboard`

Admin:
- Dashboard
- Master Data: Guru, Siswa, Program, Level, Group, Mata Pelajaran
- Bank Soal
- Ujian: Jadwal Ujian, Token, Monitor Ujian, Monitor Peserta, Reset Login
  - plus Absensi Peserta
- Evaluasi / Hasil Nilai
- Cetak
- Config / Settings

Guru:
- Dashboard
- Bank Soal
- Ujian: Jadwal Ujian, Token, Monitor Ujian, Monitor Peserta, Reset Login (opsional)
  - plus Absensi Peserta
- Evaluasi
- Profil Guru

Catatan akses Admin vs Guru:
- Untuk `Bank Soal` dan `Ujian`, admin dan guru memakai layar/form yang sama agar alur operasional konsisten.
- Admin dapat bekerja lintas guru, termasuk memilih `Guru` saat membuat ujian dan menyalin bank soal ke guru lain.
- Guru tidak melihat field pemilihan guru dan hanya bisa mengakses bank soal, ujian, token, monitor, serta evaluasi yang terkait dengan kepemilikannya sendiri.

Siswa:
- Dashboard
- Daftar Ujian / Ruang Ujian
- Kerjakan Ujian
- Hasil Ujian
- Informasi / Pengumuman (opsional)
- Absensi (opsional)
- Profil Saya

## Status Frontend

Halaman yang sudah punya isi dan terhubung ke backend:

- Login JWT ke `/api/v1/auth/login` dan role UI mengikuti hasil `/api/v1/me`
- Dashboard Admin dengan statistik backend
- Admin Master Data:
  - `Guru`
  - `Siswa`
  - `Verifikasi Pendaftaran`
  - `Program`
  - `Level`
  - `Group`
  - `Mata Pelajaran`
  - `Sesi` (Master Sesi dengan jam mulai/selesai)
- Admin/Guru:
  - `Bank Soal`
  - `Jadwal Ujian`
  - `Token`
  - `Monitor Ujian`
  - `Monitor Peserta`
- Siswa:
  - `Dashboard`
  - `Daftar Ujian / Ruang Ujian`
  - `Kerjakan Ujian` (join token + pengerjaan soal)
  - `Hasil Ujian` (terhubung backend)
  - `Informasi / Pengumuman`
  - `Absensi`
  - `Profil Saya`

Aksi yang sudah tersedia di UI:

- Tambah/edit/hapus `Guru`
- Tambah/edit/hapus `Siswa`
- Approve/reject `Verifikasi Pendaftaran`
- CRUD dasar `Program`, `Level`, `Group`, `Mata Pelajaran`
- Edit set soal dan edit/hapus pertanyaan pada `Bank Soal`
- Preview/import soal dari file `.docx`
- Tambah `Jadwal Ujian`
- Attach `question set` ke ujian
- Atur `target ujian` (level/group/siswa)
- Generate dan aktif/nonaktif `Token`
- Monitor ujian dan monitor peserta dengan data operasional frontend (semi-live; menunggu realtime backend)
- Area siswa utama (`Dashboard`, `Daftar Ujian`, `Hasil Ujian`) sudah memakai data backend
- Modul `Informasi / Pengumuman` dan `Absensi` sudah terhubung backend

Halaman lain yang backend-nya belum matang sudah diubah dari kartu kosong menjadi placeholder informatif agar tidak tampil blank.

Pembersihan template frontend yang sudah dilakukan:

- Menghapus branding placeholder template seperti `One`, `John Doe`, dan search bar demo navbar.
- Mengganti identitas navbar/avatar agar mengikuti user login JWT aktif.
- Menghapus bootstrap request sample data bawaan template.
- Menghapus file demo/frontend sample yang tidak dipakai lagi (`UiView`, `UsersView`, `SettingsView`, sample cards/tables/charts, dan komponen branding template terkait).
