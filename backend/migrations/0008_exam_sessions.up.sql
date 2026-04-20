-- Exam runtime: sessions, assembled questions, answers, and events (monitor/audit).

CREATE TABLE IF NOT EXISTS exam_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  exam_id uuid NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
  student_id uuid NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  status text NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress','submitted','forced','expired')),
  started_at timestamptz NOT NULL DEFAULT now(),
  finished_at timestamptz NULL,
  client_ip inet NULL,
  user_agent text NULL,
  last_seen_at timestamptz NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (exam_id, student_id)
);

CREATE INDEX IF NOT EXISTS idx_exam_sessions_exam ON exam_sessions(exam_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_exam_sessions_student ON exam_sessions(student_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_exam_sessions_status ON exam_sessions(status, created_at DESC);

-- Freeze the assembled question order per session.
CREATE TABLE IF NOT EXISTS exam_session_questions (
  exam_session_id uuid NOT NULL REFERENCES exam_sessions(id) ON DELETE CASCADE,
  question_id uuid NOT NULL REFERENCES questions(id),
  order_no int NOT NULL CHECK (order_no > 0),
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (exam_session_id, question_id),
  UNIQUE (exam_session_id, order_no)
);

CREATE INDEX IF NOT EXISTS idx_exam_session_questions_session ON exam_session_questions(exam_session_id, order_no);

CREATE TABLE IF NOT EXISTS exam_attempts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  exam_session_id uuid NOT NULL REFERENCES exam_sessions(id) ON DELETE CASCADE,
  question_id uuid NOT NULL REFERENCES questions(id),
  answer_json jsonb NOT NULL,
  answered_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (exam_session_id, question_id)
);

CREATE INDEX IF NOT EXISTS idx_exam_attempts_session ON exam_attempts(exam_session_id, answered_at DESC);

CREATE TABLE IF NOT EXISTS exam_events (
  id bigserial PRIMARY KEY,
  exam_session_id uuid NOT NULL REFERENCES exam_sessions(id) ON DELETE CASCADE,
  type text NOT NULL,
  payload_json jsonb NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_exam_events_session ON exam_events(exam_session_id, created_at DESC);

