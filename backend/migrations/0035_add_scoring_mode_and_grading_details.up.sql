ALTER TABLE exams
ADD COLUMN IF NOT EXISTS scoring_mode text NOT NULL DEFAULT 'partial'
CHECK (scoring_mode IN ('partial', 'absolute'));

ALTER TABLE exam_sessions
ADD COLUMN IF NOT EXISTS grading_detail_json jsonb NOT NULL DEFAULT '[]'::jsonb;

