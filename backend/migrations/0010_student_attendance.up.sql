CREATE TABLE IF NOT EXISTS student_attendance (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  exam_id uuid NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
  student_id uuid NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  note text NULL,
  attended_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (exam_id, student_id)
);

CREATE INDEX IF NOT EXISTS idx_student_attendance_student_time
  ON student_attendance (student_id, attended_at DESC);

CREATE INDEX IF NOT EXISTS idx_student_attendance_exam
  ON student_attendance (exam_id, attended_at DESC);
