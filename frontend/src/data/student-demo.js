export const studentAnnouncements = [
  {
    id: 'ann-1',
    title: 'Simulasi CBT Mingguan',
    body: 'Simulasi CBT dibuka setiap Senin pukul 08.00 WIB. Gunakan akun masing-masing dan pastikan perangkat stabil.',
    category: 'Akademik',
    published_at: '2026-04-13 07:30 WIB',
  },
  {
    id: 'ann-2',
    title: 'Persiapan Ujian Biologi',
    body: 'Bawa kartu ujian digital dan hadir 15 menit sebelum ujian dimulai untuk proses verifikasi kehadiran.',
    category: 'Ujian',
    published_at: '2026-04-12 14:10 WIB',
  },
  {
    id: 'ann-3',
    title: 'Perubahan Jadwal Bahasa Indonesia',
    body: 'Jadwal ujian Bahasa Indonesia dipindahkan ke Jumat pukul 09.30 WIB karena penyesuaian ruang laboratorium.',
    category: 'Jadwal',
    published_at: '2026-04-11 16:45 WIB',
  },
]

export const studentExamCatalog = [
  {
    id: 'exam-demo-1',
    title: 'Matematika Wajib - Ulangan Harian 3',
    subject: 'Matematika',
    teacher: 'Dian Pertiwi, S.Pd',
    starts_at: '2026-04-15T08:00:00Z',
    ends_at: '2026-04-15T09:30:00Z',
    duration_minutes: 90,
    token_required: true,
    room: 'Lab CBT 1',
    status: 'upcoming',
  },
  {
    id: 'exam-demo-2',
    title: 'Biologi - Latihan AKM',
    subject: 'Biologi',
    teacher: 'Seno Adi Nugroho, M.Pd',
    starts_at: '2026-04-16T01:30:00Z',
    ends_at: '2026-04-16T03:00:00Z',
    duration_minutes: 90,
    token_required: false,
    room: 'Ruang Kelas XI IPA 1',
    status: 'upcoming',
  },
  {
    id: 'exam-demo-3',
    title: 'Bahasa Indonesia - Penilaian Tengah Semester',
    subject: 'Bahasa Indonesia',
    teacher: 'Rina Kusuma, S.Pd',
    starts_at: '2026-04-10T08:00:00Z',
    ends_at: '2026-04-10T09:00:00Z',
    duration_minutes: 60,
    token_required: true,
    room: 'Lab CBT 2',
    status: 'completed',
    score: 84,
  },
]

export const studentExamQuestions = [
  {
    id: 'q-1',
    type: 'mc_single',
    stem: 'Hasil dari 12 x 8 adalah ...',
    options: [
      { label: 'A', content: '88' },
      { label: 'B', content: '96' },
      { label: 'C', content: '108' },
      { label: 'D', content: '112' },
    ],
  },
  {
    id: 'q-2',
    type: 'true_false',
    stem: 'Pernyataan: Sel adalah unit struktural terkecil makhluk hidup.',
    options: [
      { label: 'A', content: 'Benar' },
      { label: 'B', content: 'Salah' },
    ],
  },
  {
    id: 'q-3',
    type: 'short_answer',
    stem: 'Tuliskan rumus luas lingkaran.',
    options: [],
  },
]

export const studentResultHistory = [
  {
    id: 'res-1',
    exam_title: 'Bahasa Indonesia - Penilaian Tengah Semester',
    subject: 'Bahasa Indonesia',
    submitted_at: '2026-04-10 09:02 WIB',
    score: 84,
    correct_count: 34,
    total_questions: 40,
    status: 'Tuntas',
  },
  {
    id: 'res-2',
    exam_title: 'Matematika - Quiz Fungsi',
    subject: 'Matematika',
    submitted_at: '2026-04-05 10:11 WIB',
    score: 78,
    correct_count: 21,
    total_questions: 30,
    status: 'Remedial',
  },
]
