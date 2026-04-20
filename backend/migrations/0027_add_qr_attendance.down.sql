ALTER TABLE student_attendance DROP COLUMN IF EXISTS attendance_session_id;
ALTER TABLE student_attendance DROP COLUMN IF EXISTS lat;
ALTER TABLE student_attendance DROP COLUMN IF EXISTS lon;
ALTER TABLE student_attendance DROP COLUMN IF EXISTS accuracy;
ALTER TABLE student_attendance DROP COLUMN IF EXISTS is_qr;

DROP TABLE IF EXISTS attendance_sessions;
