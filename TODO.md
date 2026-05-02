# AtigaCBT Implementation Progress

## Project Overview
AtigaCBT is a premium computer-based testing platform designed for stability, rich aesthetics, and ease of use.

## Modul Implementation Status

### ✅ Selesai (Completed)
- [x] **Arsitektur Dasar** - Gin backend, Vue frontend, PostgreSQL database.
- [x] **Manajemen User** - Multi-role (Admin, Teacher, Student) dengan Auth JWT.
- [x] **Bank Soal (M1-M3)** - Wizard buat soal, support multiple types (MC, Multiple MC, Essay, T/F, Matching).
- [x] **Quick Add Soal (Batch)** - Kartu + modal "Tambah Banyak" di editor Bank Soal untuk membuat banyak soal sekaligus (pilih tipe + jumlah, template valid, max 100/sekali).
- [x] **Log Aktivitas Login (Admin)** - Submenu Config untuk melihat siapa saja yang login + aksi hapus per log / hapus semua.
- [x] **Audit Log UI (Admin)** - Submenu Config untuk melihat riwayat aksi mutasi (POST/PUT/PATCH/DELETE) oleh admin/guru + detail payload.
- [x] **Jadwal Ujian (M4)** - Penjadwalan ujian, target peserta (level/group/siswa), kustomisasi acak soal/opsi.
- [x] **Token Ujian** - Sistem token dinamis dengan validitas waktu.
- [x] **Ruang Ujian Siswa** - Workspace ujian, navigasi, auto-save, sisa waktu.
- [x] **Real-time Monitoring (M5)** - Dashboard pengawas via SSE & Polling fallback.
- [x] **Evaluasi & Hasil (M6)** - Rekap nilai, analisis butir soal, distribusi nilai.
- [x] **Export Nilai** - Download Excel dengan analisis mendalam (sheets: Results, Item Analysis, etc).
- [x] **Cetak (M7)** - Cetak kartu ujian (multi-template), daftar hadir, dan laporan nilai.
- [x] **Validasi Sesi Ujian (M4 Refinement)** - Enforce waktu sesi (Pagi 07:00-10:00, dll) pada join logic.
- [x] **Teacher Roles & Groups Refinement** - Membatasi guru hanya bisa melihat/mengelola group yang ditugaskan.
- [x] **UI/UX Polishing** - Memperbaiki interface input UUID di Target Ujian menjadi selection.
- [x] **Analisis Subjektif (Essay)** - Interface untuk guru memberi nilai pada soal essay dan feedback.
- [x] **Notifikasi & Pengumuman** - Sistem broadcast pengumuman premium & Notification Bell untuk siswa.
- [x] **Backup & Restore** - Sistem ekspor/impor database SQL via panel admin.
- [x] **Integrasi LMS & Data Portability** - Ekspor Roster Siswa & Hasil Ujian ke CSV / JSON kompatibel LMS (Google Classroom, Moodle, Dapodik). Panel Admin `/admin/lms` dengan panduan integrasi per platform.
- [x] **Advanced Analytics Dashboard** - Visualisasi tren performa (Chart.js), analisis per mata pelajaran, dan perbandingan antar rombel/group. Insight berbasis data untuk evaluasi kurikulum. `/admin/analytics`.
- [x] **Admin Panel Smoke Audit Script** - Skrip `scripts/audit_admin.sh` untuk smoke-test endpoint yang dipakai panel admin (login, settings, analytics, LMS, master data CRUD, bank soal, jadwal ujian, token) dengan create data dummy + cleanup.
- [x] **Fix Analytics & LMS Queries (Postgres)** - Memperbaiki query yang mengacu ke kolom yang tidak ada (mis. `es.score`, `e.start_at`) dengan menghitung score via scoring engine dan memakai kolom schema yang benar (`starts_at/ends_at`, join `exam_sessions -> students -> users`).
- [x] **Template Import DOCX + LaTeX** - Template import DOCX bisa di-download dari editor Bank Soal (kartu template warna ungu/purple), mendukung penulisan LaTeX sebagai teks (`$...$` dan `$$...$$`).
- [x] **Stabilisasi Sigma/LaTeX di Editor Bank Soal** - Fitur Sigma di `RichEditor` kini menyisipkan formula sebagai MathML (bukan teks literal), konfigurasi TinyMCE di-whitelist untuk elemen MathML agar tidak tersanitasi, dan warning sidebar `vSlot` sudah diperbaiki pada komponen menu.
- [x] **Notifikasi Email/WA (M8)** - Blast pengumuman & hasil nilai via email / WhatsApp API. Sistem konfigurasi SMTP & WA di Dashboard Admin.
- [x] **LTI 1.3 Deep Integration** - Full LTI provider (OIDC & Launch) dengan otomatisasi provisioning student, LTI Deep Linking resource picker untuk guru, dan admin management UI. `/admin/lti`.
- [x] **QR Code Absensi** - Sistem absensi berbasis QR code dengan geo-fencing (Lat/Lon/Radius). `/admin/ujian/absensi/qr`.
- [x] **AI-Powered Item Suggester** - Rekomendasi perbaikan soal berdasarkan item analysis (D-Index, P-Value, Distractor strength).
- [x] **Smart Exam Delivery (Phase 1)** - Optimasi mass-join dengan binary COPY protocol, cheat detection (focus loss tracking), dan lazy-load assets.
- [x] **Pendaftaran & Registrasi Mandiri** - Sistem registrasi 3-langkah (Wizard) dengan tema premium, dukungan isian manual, verifikasi admin, dan auto-link ke akun Google saat login pertama kali. Capturing data NISN, Tanggal Lahir, Sekolah, dan Tahun Ajaran. `/auth/google/register`.

- [x] **Simplified Run Script** - Skrip `run-local.sh` untuk memudahkan menjalankan seluruh sistem dengan satu perintah.
- [x] **Stability & Reliability Refinement** - Perbaikan menyeluruh pada alur pengerjaan siswa: perbaikan sinkronisasi token (Lateral Join), standarisasi zona waktu Indonesia (WIB/WITA/WIT), optimasi API Workspace Lembar Ujian, dan navigasi (router) ruang ujian.
- [x] **Timezone-Aware Session Windows** - Implementasi validasi jendela waktu sesi (Pagi, Siang, dll) yang akurat berdasarkan zona waktu sistem (Asia/Jakarta) pada saat Join ujian.
- [x] **Modern Exam Management UI** - Pembaruan antarmuka pengelola ujian dengan pewarnaan status yang intuitif (Draft, Archive) dan fitur hapus jadwal ujian langsung dari dashboard.
- [x] **Final Polish & Documentation** - Membersihkan kode, dokumentasi API, dan panduan penggunaan.
- [x] **Perbaikan Approval Pendaftaran Siswa** - Mengatasi error "nis required" saat admin approve pendaftaran dengan implementasi *Triple Fallback Logic* (NIS → NISN → Username) pada handler `ApproveRegistration` dan `PatchRegistration`; logika lookup (level/group/program) dibuat toleran agar tidak memblokir approval saat nama tidak persis cocok; form `GoogleRegistrationForm.vue` diperbarui dengan kolom NIS eksplisit.
- [x] **UI Tema Warna Konsisten (Admin)** - Pembaruan estetika pada halaman-halaman admin: Verifikasi Pendaftaran (kolom aksi ungu/purple), Log Aktivitas Login (Refresh & Apply biru, Reset hijau, Hapus >30 Hari ungu), Audit Log (Export CSV hijau solid, Hapus >30 Hari ungu, Reset hijau solid, Refresh & Apply biru).
- [x] **Paritas Panel Admin/Guru untuk Bank Soal & Ujian** - Panel admin dan guru kini memakai view operasional yang sama untuk Bank Soal, Import Soal, Pratinjau, Jadwal Ujian, Token, Monitor, dan Evaluasi; perbedaannya dipertahankan di level role/scope: guru hanya mengelola data miliknya sendiri, admin bisa memilih guru saat membuat ujian dan menyalin bank soal ke guru lain.
- [x] **Refinement Import Soal & LaTeX Short Answer** - Tombol template soal dipusatkan hanya di submenu Impor Soal, editor bank soal otomatis membuka soal nomor 1 saat set sudah ada, dan alur LaTeX untuk isian singkat diperbaiki agar formula tetap stabil di editor serta kunci jawaban dirender dengan benar di halaman pratinjau.
- [x] **Bulk Approve Verifikasi Pendaftaran** - Admin kini dapat menyetujui pendaftaran `pending` secara massal berdasarkan filter aktif (role dan pencarian), dengan ringkasan hasil (approved/gagal/sisa) agar penanganan pendaftar campuran siswa-guru lebih cepat.


### 📊 Future Roadmap
- [ ] **LTI Advantage Services** - LTI Assignment & Grades Service (AGS) support.
- [ ] **Role Capability Matrix** - Dokumentasi matriks hak akses Admin vs Guru vs Siswa per menu agar perubahan scope fitur tidak membingungkan saat UI tetap memakai view yang sama.

---
**Current Focus:** Ready for Production.

## Production Deployment Notes (VPS)

- VPS `107.173.21.147` menggunakan **Dokploy + Traefik**.
- Port `80/443` sudah dipegang container `dokploy-traefik` (`docker-proxy`), sehingga **Nginx host tidak bisa bind** ke port tersebut.
- Keputusan deployment:
- **Jangan gunakan Nginx host** sebagai reverse proxy utama selama Dokploy aktif.
- Routing domain `cbt.aantriono.sch.id` dilakukan melalui **Traefik (Dokploy)**.
- Backend AtigaCBT tetap jalan pada port internal (mis. `127.0.0.1:8080` / service internal container), lalu diproxy oleh Traefik.
- Frontend dilayani sebagai service/app di Dokploy dan domain diarahkan lewat Traefik.
- Jika ingin memaksa Nginx host, `dokploy-traefik` harus dimatikan (tidak direkomendasikan karena memutus routing Dokploy).

## Cara Login (Dev)

Jika login selalu muncul pesan **network error**, biasanya karena frontend tidak bisa reach backend (API belum jalan atau base URL salah).

1) Pastikan backend hidup:

```bash
docker compose up -d
curl -sS http://127.0.0.1:8080/healthz
```

2) Jalankan backend + frontend (cara termudah):

```bash
./run-local.sh
```

3) Buka UI di:

- `http://127.0.0.1:5173/admin-one-vue-tailwind/`

4) Login default (setelah seed admin):

- Username: `admin`
- Password: `admin12345`

Catatan (DB baru): jalankan migrate + seed sekali dari folder `backend`:

```bash
cd backend
export DATABASE_URL="postgres://atigacbt:atigacbt@localhost:5433/atigacbt?sslmode=disable"
export JWT_SECRET="7f59f6b9c9f2b8e8a8b8c8d8e8f808182838485868788898a8b8c8d8e8f8081"
../.tooling/go/bin/go run ./cmd/migrate
export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=admin12345
../.tooling/go/bin/go run ./cmd/seed
```

Jika frontend diakses dari device lain, set `frontend/.env`:

- `VITE_API_BASE_URL=http://IP-SERVER:8080`
