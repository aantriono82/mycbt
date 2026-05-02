# Template Soal (Untuk Import DOCX)

Backend import DOCX membaca **teks** dan mengenali pola tertentu. Template yang di-download dari UI sekarang sudah berupa file `.docx`, jadi bisa langsung dibuka, diisi, lalu di-import kembali.

## Aturan Umum

- Setiap soal harus diawali nomor: `1.` atau `1)`
- (Disarankan) tulis tipe soal dengan tag di dalam kurung siku: `[mc_single]`, `[mc_multiple]`, `[true_false]`, `[short_answer]`, `[matching]`, `[essay]`
- Opsi PG: `A. ...` sampai `F. ...`
- Kunci jawaban: `Answer: ...` atau `Kunci: ...`
- Menjodohkan: gunakan `kiri => kanan` per baris
- Isian singkat: pisahkan jawaban alternatif dengan `|`
- LaTeX: tulis sebagai teks biasa, gunakan delimiter umum:
  - Inline: `$...$`
  - Blok: `$$...$$`

Catatan:
- Jika tag tipe tidak ditulis:
  - Jika ada opsi `A.`/`B.` maka default `mc_single`
  - Jika tidak ada opsi maka default `essay`
- Gambar/diagram di DOCX belum ikut terimport (yang dibaca hanya teks).

## Contoh Lengkap Semua Tipe

Lihat file: `frontend/public/templates/template-soal-docx.docx` (bisa langsung di-download dari UI).
