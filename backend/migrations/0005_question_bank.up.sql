CREATE TABLE IF NOT EXISTS question_sets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  subject_id uuid NOT NULL REFERENCES subjects(id),
  owner_teacher_id uuid NULL REFERENCES teachers(id),
  title text NOT NULL,
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_question_sets_subject ON question_sets(subject_id);
CREATE INDEX IF NOT EXISTS idx_question_sets_owner_teacher ON question_sets(owner_teacher_id);

CREATE TABLE IF NOT EXISTS questions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  question_set_id uuid NOT NULL REFERENCES question_sets(id) ON DELETE CASCADE,
  type text NOT NULL CHECK (type IN ('mc_single','mc_multiple','matching','short_answer','essay','true_false')),
  stem text NOT NULL,
  order_no int NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_questions_set ON questions(question_set_id);

CREATE TABLE IF NOT EXISTS question_options (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id uuid NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  label text NOT NULL,
  content text NOT NULL,
  is_correct boolean NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_question_options_question ON question_options(question_id);

