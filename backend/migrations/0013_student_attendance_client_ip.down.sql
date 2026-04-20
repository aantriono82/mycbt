DROP INDEX IF EXISTS idx_student_attendance_client_ip;

ALTER TABLE student_attendance
  DROP COLUMN IF EXISTS client_ip;
