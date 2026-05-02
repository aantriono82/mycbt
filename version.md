# Script Dump & Bump Versi Aplikasi

File script: `scripts/version.sh`  
Target versi yang dikelola: `frontend/package.json`

## Cara Menjalankan

Jalankan dari root project (`/home/aantriono/dev/atigacbt`):

```bash
./scripts/version.sh dump
```

Output contoh:

```text
frontend/package.json version: 4.1.0
```

## Bump Versi Otomatis (SemVer)

### 1) Patch (x.y.Z)

```bash
./scripts/version.sh bump patch
```

### 2) Minor (x.Y.0)

```bash
./scripts/version.sh bump minor
```

### 3) Major (X.0.0)

```bash
./scripts/version.sh bump major
```

## Set Versi Manual

```bash
./scripts/version.sh bump set 4.2.0
```

## Catatan

- Script memakai `npm version --no-git-tag-version`, jadi:
  - update `frontend/package.json`
  - update `frontend/package-lock.json` (jika ada)
  - tidak membuat git tag otomatis
- Pastikan `node` dan `npm` tersedia di environment.
