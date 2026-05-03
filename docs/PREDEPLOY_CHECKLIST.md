# Predeploy Checklist

## Sebelum staging/production

- [ ] `backend/.env` tidak lagi tersimpan di git
- [ ] `backend/uploads/*` tidak lagi tersimpan di git
- [ ] seluruh secret production baru sudah dibuat:
  - [ ] `JWT_SECRET`
  - [ ] `POSTGRES_PASSWORD`
  - [ ] `REDIS_PASSWORD`
  - [ ] OAuth / SMTP secret bila fitur dipakai
- [ ] `deploy/.env.production` sudah diisi domain final
- [ ] `APP_URL`, `CORS_ORIGINS`, dan `GOOGLE_REDIRECT_URL` konsisten
- [ ] strategi upload dipilih:
  - [ ] local volume
  - [ ] RustFS / S3-compatible

## Verifikasi aplikasi

- [ ] `make test-backend`
- [ ] `cd frontend && npm test`
- [ ] `cd frontend && npm run build`
- [ ] image backend bisa build
- [ ] image frontend bisa build
- [ ] migrasi jalan dari container
- [ ] seed admin jalan dari container
- [ ] smoke test lolos:
  - [ ] `./scripts/audit_admin.sh`
  - [ ] `./scripts/audit_teacher_full.sh`
  - [ ] `./scripts/audit_student_full.sh`

## Verifikasi operasional

- [ ] backup PostgreSQL sudah dijadwalkan
- [ ] backup upload sudah dijadwalkan jika pakai storage lokal
- [ ] hanya service `web` yang diekspos publik
- [ ] TLS/domain sudah aktif
- [ ] log dan healthcheck bisa dipantau
- [ ] kredensial admin awal diganti setelah seed pertama
