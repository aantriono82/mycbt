-- Attendance Sessions for QR Code based attendance
CREATE TABLE IF NOT EXISTS attendance_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    exam_id uuid NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    token text NOT NULL UNIQUE,
    lat double precision,
    lon double precision,
    radius_meters int DEFAULT 50,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Update student_attendance to support QR sessions and geolocation
ALTER TABLE student_attendance ADD COLUMN IF NOT EXISTS attendance_session_id uuid REFERENCES attendance_sessions(id) ON DELETE SET NULL;
ALTER TABLE student_attendance ADD COLUMN IF NOT EXISTS lat double precision;
ALTER TABLE student_attendance ADD COLUMN IF NOT EXISTS lon double precision;
ALTER TABLE student_attendance ADD COLUMN IF NOT EXISTS accuracy double precision;
ALTER TABLE student_attendance ADD COLUMN IF NOT EXISTS is_qr boolean DEFAULT false;
