ALTER TABLE exam_sessions
DROP COLUMN IF EXISTS grading_detail_json;

ALTER TABLE exams
DROP COLUMN IF EXISTS scoring_mode;

