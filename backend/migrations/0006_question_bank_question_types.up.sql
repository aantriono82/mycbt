-- Add per-type tables for questions beyond multiple choice.

CREATE TABLE IF NOT EXISTS question_matching_pairs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id uuid NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  left_content text NOT NULL,
  right_content text NOT NULL,
  order_no int NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_question_matching_pairs_question ON question_matching_pairs(question_id);

CREATE TABLE IF NOT EXISTS question_short_answers (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id uuid NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  answer_text text NOT NULL,
  order_no int NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_question_short_answers_question ON question_short_answers(question_id);

-- Exactly one row per true/false question.
CREATE TABLE IF NOT EXISTS question_true_false (
  question_id uuid PRIMARY KEY REFERENCES questions(id) ON DELETE CASCADE,
  correct bool NOT NULL
);

-- Exactly one row per essay question.
CREATE TABLE IF NOT EXISTS question_essays (
  question_id uuid PRIMARY KEY REFERENCES questions(id) ON DELETE CASCADE,
  rubric_text text NULL,
  max_score int NULL CHECK (max_score IS NULL OR max_score >= 0)
);

