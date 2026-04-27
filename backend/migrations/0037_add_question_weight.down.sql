ALTER TABLE questions
DROP CONSTRAINT IF EXISTS chk_questions_weight_positive;

ALTER TABLE questions
DROP COLUMN IF EXISTS weight;

