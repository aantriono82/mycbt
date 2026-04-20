ALTER TABLE teachers ADD COLUMN jenjang text NULL;

CREATE TABLE IF NOT EXISTS teacher_levels (
  teacher_id uuid NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
  level_id uuid NOT NULL REFERENCES levels(id) ON DELETE CASCADE,
  PRIMARY KEY (teacher_id, level_id)
);
