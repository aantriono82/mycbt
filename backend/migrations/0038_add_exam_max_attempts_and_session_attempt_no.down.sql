DROP INDEX IF EXISTS uniq_exam_sessions_exam_student_attempt;

ALTER TABLE exam_sessions
  ADD CONSTRAINT exam_sessions_exam_id_student_id_key UNIQUE (exam_id, student_id);

ALTER TABLE exam_sessions
  DROP CONSTRAINT IF EXISTS exam_sessions_attempt_no_positive;

ALTER TABLE exam_sessions
  DROP COLUMN IF EXISTS attempt_no;

ALTER TABLE exams
  DROP COLUMN IF EXISTS max_attempts;
