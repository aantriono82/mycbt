CREATE TABLE IF NOT EXISTS student_exam_dismissals (
  student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  exam_id UUID NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
  dismissed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (student_id, exam_id)
);

CREATE INDEX IF NOT EXISTS idx_student_exam_dismissals_exam_id
  ON student_exam_dismissals (exam_id);
