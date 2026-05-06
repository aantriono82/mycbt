# Deploy AtigaCBT ke Dokploy

Panduan ini mengikuti arsitektur yang sudah ada di repo:

- `web`: frontend Vue yang dibuild lalu dilayani Nginx
- `api`: backend Go (Gin)
- `db`: PostgreSQL
- `redis`: Redis
- `uploads_data`: volume upload lokal bila `UPLOAD_PROVIDER=local`

File deploy yang dipakai:

- `deploy/compose.production.yml`
- `deploy/.env.production.example`

## 1. Prasyarat

Sebelum deploy, siapkan:

- VPS sudah terpasang Dokploy
- domain/subdomain yang mengarah ke VPS, mis. `cbt.sekolah.sch.id`
- akses repo Git yang dipakai Dokploy
- secret production:
  - password PostgreSQL
  - password Redis
  - `JWT_SECRET` dari `openssl rand -hex 32`
  - password admin awal

## 2. Arsitektur yang dipakai di Dokploy

Untuk repo ini, pola paling aman adalah:

- expose hanya service `web`
- service `api`, `db`, dan `redis` tetap internal
- akses publik memakai satu domain, mis. `https://cbt.sekolah.sch.id`
- Nginx di container `web` akan meneruskan:
  - `/api/*` ke `api:8080`
  - `/uploads/*` ke `api:8080`
  - `/healthz` ke `api:8080`

Dengan pola ini:

- frontend tidak perlu bicara ke host/port internal
- CORS lebih sederhana
- image frontend yang dibuild tetap konsisten

## 3. Buat project di Dokploy

Di Dokploy:

1. `Create Project`
2. pilih source dari Git repository repo ini
3. pilih tipe `Docker Compose`
4. untuk compose file, isi:

```text
deploy/compose.production.yml
```

5. branch: gunakan branch production Anda, biasanya `main`

## 4. Tambahkan environment variables

Ambil template dari:

`deploy/.env.production.example`

Lalu masukkan ke Dokploy pada bagian environment. Nilai minimum yang harus diisi:

```env
WEB_PORT=8088

POSTGRES_DB=atigacbt
POSTGRES_USER=atigacbt
POSTGRES_PASSWORD=GANTI_PASSWORD_DB
DATABASE_URL=postgres://atigacbt:GANTI_PASSWORD_DB@db:5432/atigacbt?sslmode=disable

REDIS_PASSWORD=GANTI_PASSWORD_REDIS
REDIS_DB=0
REDIS_PREFIX=atigacbt

APP_URL=https://cbt.sekolah.sch.id
CORS_ORIGINS=https://cbt.sekolah.sch.id

# Boleh kosong bila frontend dan backend satu domain
VITE_API_BASE_URL=

JWT_SECRET=HASIL_OPENSSL_RAND_HEX_32
JWT_ISSUER=atigacbt
JWT_TTL_MINUTES=120

UPLOAD_PROVIDER=local

ADMIN_USERNAME=admin
ADMIN_PASSWORD=GANTI_PASSWORD_ADMIN
ADMIN_NAME=Administrator
ADMIN_EMAIL=admin@sekolah.sch.id
```

Catatan penting:

- `DATABASE_URL` untuk compose bawaan ini memakai host internal `db`
- `REDIS_ADDR` tidak perlu diisi manual karena compose sudah menetapkan `redis:6379`
- `VITE_API_BASE_URL` boleh dikosongkan untuk mode satu domain; frontend akan memakai path relatif yang diproxy oleh Nginx
- jangan pakai `localhost` untuk `APP_URL`, `CORS_ORIGINS`, atau `DATABASE_URL`

## 5. Build args frontend

Service `web` memakai build arg:

```text
VITE_API_BASE_URL
```

Untuk setup satu domain, biarkan kosong. Itu memang sudah sesuai dengan komentar di `frontend/.env.example`.

Kalau Anda sengaja memisahkan frontend dan API ke domain berbeda, baru isi misalnya:

```env
VITE_API_BASE_URL=https://api.sekolah.sch.id
```

Tetapi itu bukan setup yang direkomendasikan untuk repo ini.

## 6. Konfigurasi domain di Dokploy

Pasang domain hanya ke service `web`.

Contoh:

- domain: `cbt.sekolah.sch.id`
- target container/service: `web`
- target port: `80`

Biarkan `api`, `db`, dan `redis` tanpa public domain.

## 7. Jalankan migrasi sebelum aplikasi dinaikkan

Repo ini sudah menyediakan service tools:

- `migrate`
- `seed`
- `cleanup`

Urutan yang disarankan:

1. deploy service dependency dulu: `db` dan `redis`
2. jalankan job `migrate`
3. saat first deploy saja, jalankan job `seed`
4. baru naikkan `api` dan `web`

Kalau Dokploy Anda mendukung one-off command/job per service:

- jalankan service `migrate` dengan command bawaan `./migrate`
- jalankan service `seed` dengan command bawaan `./seed`

Kalau Dokploy Anda hanya menjalankan seluruh compose sekaligus, tetap pastikan:

- `migrate` pernah dijalankan sukses sebelum user memakai aplikasi
- `seed` tidak dijalankan berulang pada deploy rutin, kecuali memang ingin membuat admin awal lagi

## 8. First deploy yang aman

Urutan praktis:

1. set semua env production di Dokploy
2. deploy compose
3. cek `db` healthy
4. cek `redis` healthy
5. jalankan `migrate`
6. jalankan `seed`
7. pastikan `api` healthy
8. buka domain publik dan cek frontend tampil

Endpoint health yang tersedia:

- `/healthz`
- `/api/v1/ping`

Karena `web` memproxy `/healthz`, pengecekan publik yang paling simpel:

```text
https://cbt.sekolah.sch.id/healthz
```

## 9. Login awal

Setelah `seed` sukses, login memakai kredensial dari:

- `ADMIN_USERNAME`
- `ADMIN_PASSWORD`

UI publik akan dilayani oleh service `web`.

## 10. Upload file

Bawaan production compose memakai:

```env
UPLOAD_PROVIDER=local
```

Artinya file upload disimpan di volume Docker:

- `uploads_data`

Ini cukup untuk awal, tetapi konsekuensinya:

- backup harus mencakup database dan volume upload
- bila container pindah node tanpa volume yang sama, file upload ikut hilang

Kalau nanti ingin lebih stabil, pindah ke object storage S3-compatible dengan mengisi:

- `RUSTFS_ENDPOINT`
- `RUSTFS_ACCESS_KEY`
- `RUSTFS_SECRET_KEY`
- `RUSTFS_BUCKET`
- `RUSTFS_PUBLIC_BASE_URL`

## 11. Verifikasi setelah deploy

Cek minimal hal berikut:

1. domain utama terbuka
2. `https://domain/healthz` mengembalikan sukses
3. login admin berhasil
4. halaman dashboard memuat data
5. request ke `/api/v1/auth/login` tidak kena CORS
6. upload file berjalan
7. file upload bisa diakses lagi lewat `/uploads/...`

Kalau ingin smoke test dari mesin lokal atau terminal lain:

```bash
BASE_URL=https://cbt.sekolah.sch.id \
ADMIN_USERNAME=admin \
ADMIN_PASSWORD='PASSWORD_ADMIN' \
./scripts/predeploy_smoke.sh
```

## 12. Hal yang paling sering salah

Masalah yang paling umum saat deploy repo ini ke Dokploy:

1. `DATABASE_URL` masih mengarah ke `localhost`
2. domain dipasang ke service `api`, bukan `web`
3. `CORS_ORIGINS` belum memakai domain production
4. `JWT_SECRET` terlalu pendek atau belum diganti
5. `seed` dijalankan berulang tanpa sadar
6. `VITE_API_BASE_URL` diisi URL yang tidak sesuai jalur reverse proxy

Jika memakai arsitektur bawaan repo ini, baseline yang benar adalah:

- domain publik mengarah ke `web`
- `DATABASE_URL` mengarah ke `db`
- backend internal mengarah ke `redis:6379`
- frontend dan backend satu domain

## 13. Rekomendasi operasional

Untuk production, saya sarankan:

- aktifkan auto deploy hanya untuk branch stabil
- backup PostgreSQL harian
- backup volume `uploads_data` bila masih memakai `UPLOAD_PROVIDER=local`
- simpan file env di secret manager, bukan di repo
- ganti password admin awal setelah login pertama

## 14. Auto redeploy dari GitHub Actions (GHCR -> Dokploy)

Repo ini punya workflow:

- `.github/workflows/deploy.yml`

Alurnya:

1. setiap push ke `main`, workflow build frontend image
2. workflow push image ke GHCR dengan tag:
   - `git-<shortsha>` (immutable)
   - `latest`
3. setelah push image sukses, workflow memanggil API Dokploy:
   - `POST /api/application.deploy`

Agar trigger Dokploy aktif, isi GitHub Secrets berikut:

- `DOKPLOY_URL`  
  contoh: `https://dokploy.domain-anda.com` (tanpa slash di akhir)
- `DOKPLOY_API_KEY`  
  buat dari profile Dokploy
- `DOKPLOY_APPLICATION_ID`  
  ID service/application yang ingin dideploy

Cara ambil `applicationId` via API Dokploy:

```bash
curl -X GET "https://dokploy.domain-anda.com/api/project.all" \
  -H "accept: application/json" \
  -H "x-api-key: <DOKPLOY_API_KEY>"
```

Lalu cari `applicationId` milik service frontend Anda.

## 15. Ringkasan singkat

Kalau ingin versi paling pendek:

1. import repo ini sebagai project Docker Compose di Dokploy
2. pakai `deploy/compose.production.yml`
3. isi env dari `deploy/.env.production.example`
4. expose hanya service `web`
5. arahkan domain ke `web:80`
6. jalankan `migrate`
7. jalankan `seed` sekali
8. verifikasi `https://domain/healthz`
