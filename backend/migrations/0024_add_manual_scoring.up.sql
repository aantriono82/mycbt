-- Add manual scoring for essays
ALTER TABLE exam_attempts ADD COLUMN manual_score INT NULL;
ALTER TABLE exam_attempts ADD COLUMN manual_feedback TEXT NULL;
