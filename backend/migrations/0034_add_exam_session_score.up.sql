ALTER TABLE exam_sessions
ADD COLUMN IF NOT EXISTS score int NULL CHECK (score >= 0 AND score <= 100);

