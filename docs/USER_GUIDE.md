# AtigaCBT User Guide

Selamat datang di panduan penggunaan platform AtigaCBT. Dokumen ini menjelaskan alur kerja utama untuk Administrator, Guru, dan Siswa.

## 1. Peran Administrator

### Dashboard & Master Data
Administrator memiliki akses penuh untuk mengelola infrastruktur data sekolah:
- **Guru & Siswa**: Mengelola akun pengguna melalu (Import Excel tersedia).
- **Registrasi**: Memverifikasi permintaan pendaftaran mandiri dari siswa.
- **Sesi**: Mengatur jadwal waktu (misalnya Sesi 1: 07:30 - 09:30).

### Pengaturan Sistem
- **General Settings**: Mengonfigurasi nama sekolah, logo, dan rincian kontak yang akan tampil di kartu ujian dan laporan.
- **Integrasi LMS**: Mengatur koneksi ke LMS eksternal (Moodle/Canvas) menggunakan standar LTI 1.3.

### Cetak Dokumen
Administrator dapat mencetak dokumen administratif massal:
- Kartu Peserta Ujian.
- Daftar Hadir Peserta.
- Berita Acara Ujian.

---

## 2. Peran Guru

### Bank Soal
Guru mengelola konten ujian melalui menu Bank Soal:
- **Buat Set Soal**: Menentukan Mapel, Jenjang, dan Level.
- **Editor Soal**: Mendukung berbagai tipe (Pilihan Ganda, PG Kompleks, Menjodohkan, Isian Singkat, dan Essay).
- **Import DOCX**: Mengunggah soal dari Microsoft Word menggunakan template standar.

### Pelaksanaan Ujian
- **Jadwal Ujian**: Membuat jadwal dengan menentukan token, durasi, dan target (Kelas/Group/Siswa tertentu).
- **Token**: Mengelola token aktif yang dibutuhkan siswa untuk masuk ke ruang ujian.
- **Monitoring**: Memantau progres siswa secara real-time, mendeteksi kecurangan (pindah tab), dan melakukan Force Submit jika diperlukan.

### Evaluasi & Hasil
- **Hasil Ujian**: Melihat nilai otomatis siswa segera setelah ujian selesai.
- **Koreksi Essay**: Melakukan penilaian manual untuk tipe soal essay (rubrik tersedia).
- **Analisis Butir Soal**: Melihat statistik kualitas soal (Daya Pembeda & Tingkat Kesukaran) serta saran perbaikan otomatis dari sistem.
- **Blast Hasil**: Mengirimkan nilai siswa ke Email atau WhatsApp (jika dikonfigurasi).

---

## 3. Peran Siswa

### Persiapan Ujian
- **Dashboard**: Melihat pengumuman terbaru dari sekolah atau guru.
- **Daftar Ujian**: Melihat jadwal ujian yang akan datang dan sedang berlangsung.

### Ruang Ujian
- **Masuk Ujian**: Siswa memasukkan token yang diberikan oleh guru pengawas.
- **Mengerjakan Soal**: Menjawab soal dengan navigasi yang intuitif. Sistem menyimpan jawaban secara otomatis setiap kali ada perubahan (autosave).
- **Integritas**: Sistem mendeteksi jika siswa mencoba membuka tab lain atau aplikasi lain.

### Pasca Ujian
- **Hasil**: Melihat rincian nilai dan analisis jawaban jika diizinkan oleh guru.
- **Absensi**: Melakukan scan QR Code absensi yang ditampilkan oleh pengawas untuk memverifikasi kehadiran di lokasi.
