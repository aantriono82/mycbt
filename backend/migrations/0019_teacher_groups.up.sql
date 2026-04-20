CREATE TABLE IF NOT EXISTS teacher_groups (
  teacher_id uuid NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
  group_id uuid NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  PRIMARY KEY (teacher_id, group_id)
);
