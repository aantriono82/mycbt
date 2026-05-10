ALTER TABLE questions
  DROP COLUMN IF EXISTS explanation;

ALTER TABLE exams
  DROP COLUMN IF EXISTS show_discussion_to_students;
