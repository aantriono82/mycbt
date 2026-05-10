ALTER TABLE exams
  ADD COLUMN IF NOT EXISTS show_discussion_to_students boolean NOT NULL DEFAULT false;

ALTER TABLE questions
  ADD COLUMN IF NOT EXISTS explanation text NOT NULL DEFAULT '';
