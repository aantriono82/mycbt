CREATE TABLE IF NOT EXISTS student_announcement_reads (
  student_id uuid NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  announcement_id uuid NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
  read_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (student_id, announcement_id)
);

CREATE INDEX IF NOT EXISTS idx_student_announcement_reads_student
  ON student_announcement_reads (student_id, read_at DESC);
