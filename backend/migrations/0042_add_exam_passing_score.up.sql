ALTER TABLE exams
  ADD COLUMN IF NOT EXISTS passing_score int NOT NULL DEFAULT 75
  CHECK (passing_score >= 0 AND passing_score <= 100);
