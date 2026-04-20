CREATE TABLE IF NOT EXISTS announcements (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  title text NOT NULL,
  body text NOT NULL,
  category text NOT NULL DEFAULT 'general',
  is_active boolean NOT NULL DEFAULT true,
  published_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NULL,
  target_level_id uuid NULL REFERENCES levels(id) ON DELETE SET NULL,
  target_group_id uuid NULL REFERENCES groups(id) ON DELETE SET NULL,
  target_student_id uuid NULL REFERENCES students(id) ON DELETE SET NULL,
  created_by_user_id uuid NULL REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_announcements_active_window
  ON announcements (is_active, published_at DESC, expires_at);

CREATE INDEX IF NOT EXISTS idx_announcements_targets
  ON announcements (target_student_id, target_group_id, target_level_id);

INSERT INTO announcements (title, body, category, is_active)
VALUES
  (
    'Selamat Datang di Portal CBT',
    'Gunakan menu Daftar Ujian untuk melihat ujian yang tersedia dan masuk dengan token resmi dari pengawas.',
    'informasi',
    true
  ),
  (
    'Panduan Ujian',
    'Pastikan perangkat stabil, koneksi internet aman, dan submit jawaban sebelum waktu ujian berakhir.',
    'pengumuman',
    true
  );
