# mycbt

Scaffold aplikasi CBT (Computer Based Test) dengan:

- Backend: Gin (Go) in `backend/`
- Frontend: Admin One Vue Tailwind in `frontend/`

Dokumen implementasi lengkap: lihat [docs/BLUEPRINT.md](/home/aantriono/dev/mycbt/docs/BLUEPRINT.md).
Kontrak API awal (OpenAPI): [docs/openapi.yaml](/home/aantriono/dev/mycbt/docs/openapi.yaml).

## Status Implementasi (2026-04-22)

- **Bulk Approve Verifikasi Pendaftaran**: Halaman `/admin/master-data/verifikasi-pendaftaran` kini memiliki aksi `Approve Semua Sesuai Filter` untuk memproses pendaftaran `pending` secara massal (filter `role` + `q`). Backend mengeksekusi approval per-item dengan alur role-aware (`student`/`teacher`) dan mengembalikan ringkasan hasil (approved, failed, remaining + failure details terbatas).
- **Paritas Panel Admin/Guru untuk Bank Soal & Ujian**: Route admin dan guru kini sengaja memakai view operasional yang sama untuk `Bank Soal`, `Impor Soal`, `Pratinjau`, `Jadwal Ujian`, `Token`, `Monitor`, dan `Evaluasi`. Perbedaannya ada di RBAC dan scope data: guru hanya melihat/mengelola miliknya sendiri, sedangkan admin punya kontrol lintas guru.
- **Refinement Import Soal & LaTeX Short Answer**: Tombol template soal kini dipusatkan hanya di submenu `Impor Soal`, editor `/bank-soal/new?id=...` otomatis membuka soal nomor 1 saat bank soal sudah memiliki isi, dan alur LaTeX untuk `short_answer` diperbaiki agar formula tidak kembali ke teks mentah setelah insert/edit Sigma. Kunci jawaban isian singkat di pratinjau juga kini dirender sebagai rumus.
- **Stabilisasi Sigma/LaTeX di Editor Bank Soal**: Tombol Sigma pada editor `/admin/bank-soal/new` kini merender formula lewat KaTeX ke MathML dan dipertahankan oleh schema TinyMCE (`custom_elements` + `extended_valid_elements`) agar tidak berubah menjadi teks literal (contoh `\sqrt{98}`). Warning Vue `Property "vSlot" was accessed during render...` pada sidebar juga sudah diperbaiki di komponen menu.
- **Resolved Student Registration Approval**: Mengatasi error "nis required" saat admin menyetujui pendaftaran siswa. Implementasi *Triple Fallback Logic* (NIS → NISN → Username) pada backend; logika lookup nama Kelas/Rombel diperbarui agar toleran terhadap ketidakcocokan nama, sehingga approval tidak terblokir. Formulir `GoogleRegistrationForm.vue` juga diperbarui untuk menangkap kolom NIS secara eksplisit.
- **Admin UI Color Theme Consistency**: Pembaruan estetika menyeluruh pada halaman-halaman panel admin. Halaman Verifikasi Pendaftaran kini memiliki kolom aksi bertema ungu (purple). Halaman Log Aktivitas dan Audit Log memiliki skema warna tombol yang konsisten: Refresh & Apply (biru), Reset (hijau solid), Hapus >30 Hari (ungu), Export CSV (hijau solid), dan Hapus Semua (merah).
- **Admin Panel Backend Smoke Audit**: Menambahkan skrip `scripts/audit_admin.sh` untuk smoke-test endpoint yang dipakai panel admin (login, settings, analytics, LMS, master data CRUD, bank soal, ujian, token) secara end-to-end (create dummy data + cleanup).
- **Fix Analytics & LMS Endpoints**: Memperbaiki endpoint yang 500 karena query mengacu ke kolom yang tidak ada (mis. `es.score`, `e.start_at`). Analytics dan export hasil LMS kini menghitung skor lewat scoring engine dan memakai schema yang benar (`starts_at/ends_at`, join `exam_sessions -> students -> users`).
- **Template Import DOCX (LaTeX-ready)**: Template import DOCX sekarang dipusatkan di submenu `/bank-soal/import` agar tidak duplikat di editor Bank Soal. Template tetap mendukung penulisan LaTeX sebagai teks dengan delimiter `$...$` dan `$$...$$`. Panduan: `docs/template-soal-docx.md`.
- **Timezone-Aware Session Management**: Validasi jendela waktu pengerjaan (Sesi 1, 2, dsb) kini sepenuhnya akurat mengikuti zona waktu lokal (WIB/WITA/WIT), mencegah akses di luar jam yang ditentukan.
- **Modernized Exam Administration UI**: Interface pengelola ujian diperbarui dengan skema warna status yang premium (Warning untuk Draft, Contrast untuk Archive) dan fitur penghapusan jadwal yang terintegrasi.
- **Robust Workspace Navigation**: Perbaikan error navigasi (router) dan reaktivitas pada lembar pengerjaan siswa, menjamin perpindahan antar soal yang mulus dan stabil.
- **Resolved Student Access Issues**: Penanganan error 500 pada alur Join ujian yang kini memberikan pesan informatif saat terjadi mismatch waktu sesi.
- **Optimized Token Delivery**: Arsitektur backend menggunakan `LEFT JOIN LATERAL` untuk menjamin token ujian terbaru selalu tampil stabil di dashboard siswa.

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

- Node.js (already in this environment)
- Go: install lokal (tanpa sudo) via:

```bash
./scripts/install-go.sh
```

## Run (dev)

Cara termudah untuk menjalankan aplikasi (DB, Backend, dan Frontend):

```bash
./run-local.sh
```

Catatan (first time / DB baru): `run-local.sh` hanya menyalakan proses. Jika tabel belum ada atau user admin belum dibuat, jalankan migrate + seed sekali:

```bash
cd backend
export DATABASE_URL="postgres://mycbt:mycbt@localhost:5433/mycbt?sslmode=disable"
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

#### 2. Backend (migrate + seed)
```bash
cd backend
export DATABASE_URL="postgres://mycbt:mycbt@localhost:5433/mycbt?sslmode=disable"
export JWT_SECRET="7f59f6b9c9f2b8e8a8b8c8d8e8f808182838485868788898a8b8c8d8e8f8081"
../.tooling/go/bin/go run ./cmd/migrate

export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=admin12345
../.tooling/go/bin/go run ./cmd/seed

../.tooling/go/bin/go run ./cmd/api
```

#### 3. Frontend
```bash
cd frontend
# optional: cp .env.example .env lalu set VITE_API_BASE_URL
npm run dev
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
  - Backend: [backend/README.md](/home/aantriono/dev/mycbt/backend/README.md)
  - Frontend: [frontend/README.md](/home/aantriono/dev/mycbt/frontend/README.md)

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
