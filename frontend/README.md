# Frontend (Vue 3 + Vite)

Frontend aplikasi CBT ini berbasis template UI `admin-one-vue-tailwind`, lalu disesuaikan untuk kebutuhan `AtigaCBT` (routing/menu per role dan halaman placeholder modul).

## Dev

```bash
cd frontend
npm ci

# optional: cp .env.example .env lalu ubah VITE_API_BASE_URL
npm run dev
```

Default: `http://localhost:5173`

## Env

- `VITE_API_BASE_URL` (default: `http://localhost:8080`)

## Rute Utama (hash mode)

- Admin: `/#/admin/dashboard`
- Guru: `/#/teacher/dashboard`
- Siswa: `/#/student/dashboard`

Catatan: role frontend mengikuti hasil login backend (`JWT` + `GET /api/v1/me`). Akses ke route role lain akan diarahkan ulang otomatis.
