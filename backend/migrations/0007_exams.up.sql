CREATE TABLE IF NOT EXISTS exams (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  subject_id uuid NOT NULL REFERENCES subjects(id),
  teacher_id uuid NOT NULL REFERENCES teachers(id),
  title text NOT NULL,
  starts_at timestamptz NOT NULL,
  ends_at timestamptz NOT NULL,
  duration_minutes int NULL CHECK (duration_minutes IS NULL OR duration_minutes > 0),
  shuffle_questions boolean NOT NULL DEFAULT false,
  shuffle_options boolean NOT NULL DEFAULT false,
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published','archived')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_exams_subject ON exams(subject_id);
CREATE INDEX IF NOT EXISTS idx_exams_teacher ON exams(teacher_id);
CREATE INDEX IF NOT EXISTS idx_exams_status ON exams(status);
CREATE INDEX IF NOT EXISTS idx_exams_time ON exams(starts_at, ends_at);

-- Optional link from exam to one or more question sets (bank soal)
CREATE TABLE IF NOT EXISTS exam_question_sets (
  exam_id uuid NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
  question_set_id uuid NOT NULL REFERENCES question_sets(id),
  num_questions int NULL CHECK (num_questions IS NULL OR num_questions > 0),
  PRIMARY KEY (exam_id, question_set_id)
);

-- Targeting: exactly one of level_id/group_id/student_id must be set.
CREATE TABLE IF NOT EXISTS exam_targets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  exam_id uuid NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
  level_id uuid NULL REFERENCES levels(id),
  group_id uuid NULL REFERENCES groups(id),
  student_id uuid NULL REFERENCES students(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK ((level_id IS NOT NULL)::int + (group_id IS NOT NULL)::int + (student_id IS NOT NULL)::int = 1)
);

CREATE INDEX IF NOT EXISTS idx_exam_targets_exam ON exam_targets(exam_id);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_exam_target_level ON exam_targets(exam_id, level_id) WHERE level_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uniq_exam_target_group ON exam_targets(exam_id, group_id) WHERE group_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uniq_exam_target_student ON exam_targets(exam_id, student_id) WHERE student_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS exam_tokens (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  exam_id uuid NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
  token text NOT NULL,
  valid_from timestamptz NULL,
  valid_to timestamptz NULL,
  is_active boolean NOT NULL DEFAULT true,
  created_by_user_id uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_exam_token_value ON exam_tokens(exam_id, token);
CREATE INDEX IF NOT EXISTS idx_exam_tokens_exam ON exam_tokens(exam_id, created_at DESC);

