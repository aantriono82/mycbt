ALTER TABLE questions
ADD COLUMN IF NOT EXISTS weight int NOT NULL DEFAULT 1;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_questions_weight_positive'
  ) THEN
    ALTER TABLE questions
    ADD CONSTRAINT chk_questions_weight_positive CHECK (weight > 0);
  END IF;
END $$;

