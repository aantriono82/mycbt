# Deployment

Dokumen ini menutup jalur deploy untuk VPS biasa maupun Dokploy dengan target paling aman: satu domain publik yang dilayani oleh container `web`, lalu `web` meneruskan `/api` dan `/uploads` ke `api` di jaringan internal Docker.

## Struktur deploy

- `web`: frontend statis + reverse proxy nginx
- `api`: backend Gin
- `db`: PostgreSQL
- `redis`: Redis untuk rate limit / cache / blocklist token
- `uploads_data`: volume untuk upload lokal jika `UPLOAD_PROVIDER=local`

Template production ada di:

- [deploy/compose.production.yml](/home/aantriono/dev/atigacbt/deploy/compose.production.yml)
- [deploy/.env.production.example](/home/aantriono/dev/atigacbt/deploy/.env.production.example)

## Alur deploy yang disarankan

1. Siapkan file env production:
   - salin `deploy/.env.production.example` menjadi `deploy/.env.production`
   - isi semua secret dan domain production
2. Jalankan database dan dependency:
   - `docker compose --env-file deploy/.env.production -f deploy/compose.production.yml up -d db redis`
3. Jalankan migrasi:
   - `docker compose --env-file deploy/.env.production -f deploy/compose.production.yml --profile tools run --rm migrate`
4. Seed admin pertama kali:
   - `docker compose --env-file deploy/.env.production -f deploy/compose.production.yml --profile tools run --rm seed`
5. Naikkan service aplikasi:
   - `docker compose --env-file deploy/.env.production -f deploy/compose.production.yml up -d web api`
6. Jalankan smoke test:
   - `BASE_URL=https://cbt.example.com ADMIN_USERNAME=admin ADMIN_PASSWORD='...' ./scripts/predeploy_smoke.sh`

## Catatan Dokploy

Jika deploy lewat Dokploy:

- import repo ini sebagai project Docker Compose
- pakai file `deploy/compose.production.yml`
- masukkan seluruh variable dari `deploy/.env.production`
- expose hanya service `web`
- service `api`, `db`, dan `redis` cukup internal
- jalankan job migrasi memakai service `migrate` sebelum promote release

## Domain dan proxy

Konfigurasi bawaan sekarang mengasumsikan:

- frontend dan backend berada di domain yang sama, mis. `https://cbt.example.com`
- nginx di container `web` akan meneruskan:
  - `/api/*` ke `api:8080`
  - `/uploads/*` ke `api:8080`
  - `/healthz` ke `api:8080`

Dengan pola ini, frontend tidak perlu hardcode `localhost` dan image yang sama bisa dipakai lintas environment selama reverse proxy-nya konsisten.

## Storage

Mode default production memakai upload lokal pada volume Docker:

- `UPLOAD_PROVIDER=local`
- data tersimpan di volume `uploads_data`

Kalau ingin object storage S3-compatible:

- set `UPLOAD_PROVIDER=rustfs`
- isi `RUSTFS_*`
- tetap biarkan route publik lewat `/uploads/*`

## Backup minimum

- PostgreSQL: backup harian `pg_dump` + retensi
- volume `uploads_data`: snapshot atau rsync rutin bila `UPLOAD_PROVIDER=local`
- simpan salinan `deploy/.env.production` di password manager, bukan di repo
