ALTER TABLE exams
  ADD COLUMN IF NOT EXISTS max_attempts int NOT NULL DEFAULT 1
  CHECK (max_attempts > 0);

ALTER TABLE exam_sessions
  ADD COLUMN IF NOT EXISTS attempt_no int;

UPDATE exam_sessions
SET attempt_no = 1
WHERE attempt_no IS NULL;

ALTER TABLE exam_sessions
  ALTER COLUMN attempt_no SET NOT NULL;

ALTER TABLE exam_sessions
  ADD CONSTRAINT exam_sessions_attempt_no_positive CHECK (attempt_no > 0);

ALTER TABLE exam_sessions
  DROP CONSTRAINT IF EXISTS exam_sessions_exam_id_student_id_key;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_exam_sessions_exam_student_attempt
  ON exam_sessions(exam_id, student_id, attempt_no);
