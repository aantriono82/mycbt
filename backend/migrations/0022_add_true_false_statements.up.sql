-- Create table for multiple true/false statements (ANBK Style)
CREATE TABLE IF NOT EXISTS question_true_false_statements (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id uuid NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  content text NOT NULL,
  correct boolean NOT NULL DEFAULT true,
  order_no int NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_q_tf_statements_question ON question_true_false_statements(question_id);
