# 🚀 AtigaCBT — Deployment Checklist untuk VPS

> Dokumen ini adalah panduan langkah demi langkah untuk men-deploy AtigaCBT
> ke VPS dari awal (fresh install). Ikuti urutan ini dengan tepat.

---

## 📋 Daftar Isi
1. [Persiapan Lokal](#1-persiapan-lokal)
2. [Persiapan VPS](#2-persiapan-vps)
3. [Instalasi Dependensi VPS](#3-instalasi-dependensi-vps)
4. [Setup Database PostgreSQL](#4-setup-database-postgresql)
5. [Setup Redis](#5-setup-redis)
6. [Konfigurasi File `.env` Backend](#6-konfigurasi-file-env-backend)
7. [Konfigurasi File `.env` Frontend](#7-konfigurasi-file-env-frontend)
8. [Build & Deploy Backend (Go)](#8-build--deploy-backend-go)
9. [Build & Deploy Frontend (Vue)](#9-build--deploy-frontend-vue)
10. [Setup Upload Storage](#10-setup-upload-storage)
11. [Setup Nginx](#11-setup-nginx)
12. [Setup SSL (HTTPS)](#12-setup-ssl-https)
13. [Jalankan Migrasi & Seed Admin](#13-jalankan-migrasi--seed-admin)
14. [Konfigurasi Systemd (Auto-restart)](#14-konfigurasi-systemd-auto-restart)
15. [Verifikasi Post-Deploy](#15-verifikasi-post-deploy)
16. [Checklist Keamanan](#16-checklist-keamanan)

---

## 1. Persiapan Lokal

### ☑ Build backend binary
```bash
cd backend
GIN_MODE=release go build -o atigacbt-api ./cmd/api
# Binary hasilnya: backend/atigacbt-api
```

### ☑ Build frontend static files
```bash
cd frontend
npm ci
npm run build
# Output: frontend/dist/
```

### ☑ Pastikan semua perubahan sudah di-push ke GitHub
```bash
git status   # harus clean
git push origin main
```

### ☑ Catat versi migrasi terbaru
```bash
ls backend/migrations/*.up.sql | tail -5
# Migrasi terakhir saat ini: 0035_xxx (cek ulang)
```

---

## 2. Persiapan VPS

### ☑ Spesifikasi minimum yang direkomendasikan
| Komponen | Minimum | Disarankan |
|---|---|---|
| CPU | 1 vCPU | 2 vCPU |
| RAM | 1 GB | 2 GB |
| Storage | 20 GB | 40 GB |
| OS | Ubuntu 22.04 LTS | Ubuntu 22.04 LTS |

### ☑ Update sistem
```bash
sudo apt update && sudo apt upgrade -y
```

### ☑ Buat user non-root untuk deploy
```bash
sudo adduser atigacbt
sudo usermod -aG sudo atigacbt
# Login sebagai user atigacbt untuk langkah berikutnya
su - atigacbt
```

### ☑ Setup SSH key (nonaktifkan password login)
```bash
# Di mesin lokal Anda:
ssh-copy-id atigacbt@IP_VPS

# Di VPS, edit /etc/ssh/sshd_config:
PasswordAuthentication no
sudo systemctl reload sshd
```

---

## 3. Instalasi Dependensi VPS

### ☑ Install Go (untuk build di VPS atau cukup copy binary)
> **Opsi A (lebih mudah):** Build binary di lokal, copy ke VPS via `scp`
> **Opsi B:** Install Go di VPS dan build langsung

**Opsi A — copy binary (direkomendasikan):**
```bash
# Di lokal:
scp backend/atigacbt-api atigacbt@IP_VPS:/home/atigacbt/app/
scp -r frontend/dist atigacbt@IP_VPS:/home/atigacbt/www/
scp -r backend/migrations atigacbt@IP_VPS:/home/atigacbt/app/
```

**Opsi B — install Go di VPS:**
```bash
wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile
go version
```

### ☑ Install tools tambahan
```bash
sudo apt install -y nginx curl git unzip
```

---

## 4. Setup Database PostgreSQL

### ☑ Install PostgreSQL 16
```bash
sudo apt install -y postgresql postgresql-contrib
sudo systemctl enable postgresql
sudo systemctl start postgresql
```

### ☑ Buat database & user untuk AtigaCBT
```bash
sudo -u postgres psql
```
```sql
-- Di dalam psql:
CREATE USER atigacbt WITH PASSWORD 'GANTI_PASSWORD_KUAT_DI_SINI';
CREATE DATABASE atigacbt OWNER atigacbt;
GRANT ALL PRIVILEGES ON DATABASE atigacbt TO atigacbt;
\q
```

### ☑ Verifikasi koneksi
```bash
psql -U atigacbt -h localhost -d atigacbt
# Harus berhasil masuk tanpa error
\q
```

### ⚠️ PENTING
- Gunakan password yang **panjang dan acak** (min. 20 karakter)
- Jangan pakai password `atigacbt` seperti di `docker-compose.yml` development!
- Simpan password di tempat yang aman (password manager)

---

## 5. Setup Redis

### ☑ Install Redis
```bash
sudo apt install -y redis-server
sudo systemctl enable redis-server
```

### ☑ Konfigurasi Redis dengan password
```bash
sudo nano /etc/redis/redis.conf
# Cari baris: # requirepass foobared
# Ganti dengan:
requirepass GANTI_REDIS_PASSWORD_KUAT
```

```bash
sudo systemctl restart redis-server
# Verifikasi:
redis-cli -a REDIS_PASSWORD ping
# Harus balasan: PONG
```

---

## 6. Konfigurasi File `.env` Backend

### ☑ Buat file konfigurasi di VPS
```bash
nano /home/atigacbt/app/.env
```

### ☑ Isi `.env` lengkap (ganti semua nilai dalam `<...>`)
```env
# ===== MODE =====
GIN_MODE=release
PORT=8080

# ===== URL APLIKASI =====
APP_URL=https://<domain-anda.com>
CORS_ORIGINS=https://<domain-anda.com>

# ===== DATABASE =====
DATABASE_URL=postgres://atigacbt:<DB_PASSWORD>@localhost:5432/atigacbt?sslmode=disable

# ===== REDIS (untuk rate limiting & token blocklist) =====
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=<REDIS_PASSWORD_KUAT>
REDIS_DB=0
REDIS_PREFIX=atigacbt

# ===== JWT (WAJIB DIISI DENGAN STRING ACAK KUAT) =====
# Generate dengan: openssl rand -hex 32
JWT_SECRET=<GANTI_DENGAN_STRING_ACAK_64_KARAKTER>
JWT_ISSUER=atigacbt
JWT_TTL_MINUTES=120

# ===== AKUN ADMIN PERTAMA =====
# Dipakai oleh: go run ./cmd/seed
ADMIN_USERNAME=<username_admin_sekolah>
ADMIN_PASSWORD=<PASSWORD_ADMIN_KUAT_MIN_12_KARAKTER>
ADMIN_NAME=Administrator
ADMIN_EMAIL=<email_admin@sekolah.sch.id>

# ===== UPLOAD FILE =====
# Pilihan: local | rustfs | s3
UPLOAD_PROVIDER=local
UPLOAD_LOCAL_DIR=/home/atigacbt/uploads

# ===== SMTP (opsional, untuk reset password via email) =====
# SMTP_HOST=smtp.gmail.com
# SMTP_PORT=587
# SMTP_USER=<email@gmail.com>
# SMTP_PASS=<app_password_gmail>
# SMTP_FROM=no-reply@sekolah.sch.id

# ===== GOOGLE OAUTH (opsional) =====
# FRONTEND_URL=https://<domain>
# GOOGLE_CLIENT_ID=
# GOOGLE_CLIENT_SECRET=
# GOOGLE_REDIRECT_URL=https://<domain>/api/v1/auth/google/callback
```

### ☑ Generate JWT_SECRET yang kuat
```bash
openssl rand -hex 32
# Copy output dan paste ke JWT_SECRET
```

### ☑ Set permission file .env
```bash
chmod 600 /home/atigacbt/app/.env
```

---

## 7. Konfigurasi File `.env` Frontend

### ☑ Buat `.env.production` di mesin lokal sebelum build
```bash
# File: frontend/.env.production
VITE_API_BASE_URL=https://<domain-anda.com>
```

### ☑ Build ulang frontend dengan env produksi
```bash
cd frontend
npm ci
npm run build
# Pastikan VITE_API_BASE_URL sudah benar sebelum build!
```

---

## 8. Build & Deploy Backend (Go)

### ☑ Struktur direktori di VPS
```
/home/atigacbt/
├── app/
│   ├── atigacbt-api          ← binary Go
│   ├── migrations/        ← folder migrasi SQL
│   └── .env               ← konfigurasi
├── uploads/               ← foto & file upload
└── www/
    └── dist/              ← frontend build
```

### ☑ Buat struktur folder
```bash
mkdir -p /home/atigacbt/app /home/atigacbt/uploads /home/atigacbt/www
```

### ☑ Copy file ke VPS (dari lokal)
```bash
# Binary
scp backend/atigacbt-api atigacbt@IP_VPS:/home/atigacbt/app/
# Folder migrasi (WAJIB ada di samping binary)
scp -r backend/migrations atigacbt@IP_VPS:/home/atigacbt/app/
# Pastikan binary bisa dieksekusi
ssh atigacbt@IP_VPS "chmod +x /home/atigacbt/app/atigacbt-api"
```

---

## 9. Build & Deploy Frontend (Vue)

### ☑ Copy hasil build frontend
```bash
scp -r frontend/dist/* atigacbt@IP_VPS:/home/atigacbt/www/
```

---

## 10. Setup Upload Storage

### ☑ Opsi A: Local Storage (paling mudah)
```bash
# Sudah diatur via UPLOAD_LOCAL_DIR=/home/atigacbt/uploads
mkdir -p /home/atigacbt/uploads
chmod 755 /home/atigacbt/uploads
```
> ⚠️ Pastikan direktori `/uploads` di-serve oleh Nginx (lihat bagian Nginx)

### ☑ Opsi B: RustFS / MinIO (untuk produksi besar)
```bash
# Install RustFS/MinIO di VPS terpisah atau sebagai service
# Isi UPLOAD_PROVIDER=rustfs di .env beserta RUSTFS_* lainnya
```

---

## 11. Setup Nginx

### ☑ Buat konfigurasi Nginx untuk AtigaCBT
```bash
sudo nano /etc/nginx/sites-available/atigacbt
```

```nginx
server {
    listen 80;
    server_name <domain-anda.com>;

    # Redirect HTTP ke HTTPS (aktifkan setelah SSL terpasang)
    # return 301 https://$host$request_uri;

    # Frontend (Vue SPA)
    root /home/atigacbt/www;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    # Backend API
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_read_timeout 120s;
        client_max_body_size 60M;  # untuk bulk photo ZIP
    }

    # Upload files (foto siswa dll)
    location /uploads/ {
        alias /home/atigacbt/uploads/;
        expires 30d;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # SSE (monitoring real-time)
    location /api/v1/exams/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Connection '';
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 3600s;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### ☑ Aktifkan konfigurasi
```bash
sudo ln -s /etc/nginx/sites-available/atigacbt /etc/nginx/sites-enabled/
sudo nginx -t        # harus: configuration file test is successful
sudo systemctl reload nginx
```

---

## 12. Setup SSL (HTTPS)

### ☑ Install Certbot
```bash
sudo apt install -y certbot python3-certbot-nginx
```

### ☑ Dapatkan sertifikat SSL gratis (Let's Encrypt)
```bash
sudo certbot --nginx -d <domain-anda.com>
# Ikuti instruksi, masukkan email, setuju TOS
```

### ☑ Aktifkan redirect HTTP → HTTPS di Nginx
```bash
# Edit /etc/nginx/sites-available/atigacbt
# Hapus komentar pada: return 301 https://$host$request_uri;
sudo systemctl reload nginx
```

### ☑ Verifikasi auto-renewal
```bash
sudo certbot renew --dry-run
# Harus sukses tanpa error
```

---

## 13. Jalankan Migrasi & Seed Admin

### ☑ Jalankan migrasi database (WAJIB pertama kali)
```bash
cd /home/atigacbt/app
DATABASE_URL="postgres://atigacbt:<DB_PASSWORD>@localhost:5432/atigacbt?sslmode=disable" \
  /home/atigacbt/app/atigacbt-api migrate
```
> **Catatan:** Atau jalankan binary migrate terpisah jika sudah di-build:
```bash
# Build migrate binary di lokal:
cd backend && go build -o atigacbt-migrate ./cmd/migrate
scp atigacbt-migrate atigacbt@IP_VPS:/home/atigacbt/app/

# Di VPS:
cd /home/atigacbt/app && ./atigacbt-migrate
```

### ☑ Seed akun admin pertama
```bash
# Build seed binary di lokal:
cd backend && go build -o atigacbt-seed ./cmd/seed
scp atigacbt-seed atigacbt@IP_VPS:/home/atigacbt/app/

# Di VPS (pastikan .env sudah berisi ADMIN_USERNAME, ADMIN_PASSWORD):
cd /home/atigacbt/app && ./atigacbt-seed
# Output: seeded admin user id=xxx username=xxx
```

### ☑ Verifikasi login admin
```bash
curl -X POST https://<domain>/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"<ADMIN_USERNAME>","password":"<ADMIN_PASSWORD>"}'
# Harus return: {"data":{"access_token":"...","user":{...}}}
```

---

## 14. Konfigurasi Systemd (Auto-restart)

### ☑ Buat service unit untuk backend
```bash
sudo nano /etc/systemd/system/atigacbt.service
```

```ini
[Unit]
Description=AtigaCBT Backend API
After=network.target postgresql.service redis-server.service

[Service]
Type=simple
User=atigacbt
WorkingDirectory=/home/atigacbt/app
EnvironmentFile=/home/atigacbt/app/.env
ExecStart=/home/atigacbt/app/atigacbt-api
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=atigacbt

[Install]
WantedBy=multi-user.target
```

### ☑ Aktifkan dan jalankan service
```bash
sudo systemctl daemon-reload
sudo systemctl enable atigacbt
sudo systemctl start atigacbt
sudo systemctl status atigacbt
# Harus: active (running)
```

### ☑ Lihat log real-time
```bash
sudo journalctl -u atigacbt -f
```

---

## 15. Verifikasi Post-Deploy

### ☑ Cek API health
```bash
curl https://<domain>/healthz
# Harus return: {"status":"ok","time":"..."}
```

### ☑ Cek frontend terbuka
```bash
curl -I https://<domain>/
# Harus return: HTTP/2 200
```

### ☑ Cek upload foto bisa diakses
```bash
# Upload foto siswa via admin, lalu akses URL-nya langsung
curl -I https://<domain>/uploads/avatars/xxx.jpg
# Harus: HTTP/2 200
```

### ☑ Cek fitur utama manual (browser)
- [ ] Login admin berhasil
- [ ] Tambah guru & siswa berhasil
- [ ] Import Excel guru/siswa berhasil
- [ ] Buat soal ujian berhasil
- [ ] Buat jadwal ujian berhasil
- [ ] Siswa bisa join ujian
- [ ] Cetak kartu ujian berhasil (PDF/print preview)
- [ ] Import foto massal via ZIP berhasil
- [ ] Monitor ujian real-time berfungsi (SSE)

---

## 16. Checklist Keamanan

### 🔒 Wajib sebelum go-live
- [ ] `GIN_MODE=release` (bukan `debug`)
- [ ] `JWT_SECRET` diisi string acak kuat (min 64 karakter hex)
- [ ] Password admin bukan default (`admin12345`) — ganti via seed atau UI
- [ ] Password PostgreSQL kuat dan unik
- [ ] Password Redis kuat dan unik
- [ ] File `.env` permission `600` (`chmod 600 /home/atigacbt/app/.env`)
- [ ] HTTPS aktif (Let's Encrypt)
- [ ] Port PostgreSQL (5432) **tidak** terbuka ke internet (hanya localhost)
- [ ] Port Redis (6379) **tidak** terbuka ke internet (hanya localhost)
- [ ] Port 8080 backend **tidak** terbuka ke internet (hanya via Nginx)
- [ ] Firewall aktif: `sudo ufw allow 22,80,443/tcp && sudo ufw enable`

### 🔒 Disarankan
- [ ] Ganti port SSH dari 22 ke port non-standar
- [ ] Fail2ban terpasang untuk proteksi brute force
- [ ] Regular backup database terjadwal (cron `pg_dump`)
- [ ] Monitoring uptime (UptimeRobot / Healthchecks.io)
- [ ] Setup log rotation untuk `/home/atigacbt/uploads`

---

## 🔄 Cara Update Aplikasi (setelah deploy pertama)

```bash
# 1. Di lokal — build binary baru
cd backend
GIN_MODE=release go build -o atigacbt-api ./cmd/api
cd ../frontend
npm run build

# 2. Copy ke VPS
scp backend/atigacbt-api atigacbt@IP_VPS:/home/atigacbt/app/atigacbt-api.new
scp -r frontend/dist/* atigacbt@IP_VPS:/home/atigacbt/www/

# 3. Di VPS — ganti binary & restart
ssh atigacbt@IP_VPS
mv /home/atigacbt/app/atigacbt-api.new /home/atigacbt/app/atigacbt-api
chmod +x /home/atigacbt/app/atigacbt-api

# 4. Jalankan migrasi baru (jika ada)
cd /home/atigacbt/app && ./atigacbt-migrate

# 5. Restart service
sudo systemctl restart atigacbt
sudo journalctl -u atigacbt -f  # pantau log
```

---

## 📦 Backup Database

```bash
# Backup manual
pg_dump -U atigacbt -h localhost atigacbt | gzip > /tmp/atigacbt_backup_$(date +%Y%m%d).sql.gz

# Cron job backup otomatis harian (jam 02:00)
crontab -e
# Tambahkan:
0 2 * * * pg_dump -U atigacbt -h localhost atigacbt | gzip > /home/atigacbt/backups/atigacbt_$(date +\%Y\%m\%d).sql.gz
```

---

*Dokumen ini dibuat otomatis berdasarkan konfigurasi sumber kode AtigaCBT.*
*Terakhir diperbarui: 2026-04-28*
