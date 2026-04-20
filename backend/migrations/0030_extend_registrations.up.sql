-- Add extra fields to registration_requests
ALTER TABLE registration_requests
ADD COLUMN IF NOT EXISTS jenjang TEXT,
ADD COLUMN IF NOT EXISTS gender TEXT,
ADD COLUMN IF NOT EXISTS birth_date DATE,
ADD COLUMN IF NOT EXISTS school_name TEXT,
ADD COLUMN IF NOT EXISTS academic_year TEXT,
ADD COLUMN IF NOT EXISTS nisn TEXT,
ADD COLUMN IF NOT EXISTS nis_sekolah TEXT;
