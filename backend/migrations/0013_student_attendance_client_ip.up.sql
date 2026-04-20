ALTER TABLE student_attendance
  ADD COLUMN IF NOT EXISTS client_ip inet NULL;

CREATE INDEX IF NOT EXISTS idx_student_attendance_client_ip
  ON student_attendance (client_ip);
