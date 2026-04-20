CREATE TABLE IF NOT EXISTS audit_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  request_id text NOT NULL,
  user_id uuid NULL REFERENCES users(id) ON DELETE SET NULL,
  role text NULL,
  method text NOT NULL,
  path text NOT NULL,
  query text NULL,
  status_code int NOT NULL,
  ip text NULL,
  user_agent text NULL,
  payload_json jsonb NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_role ON audit_logs(role);
CREATE INDEX IF NOT EXISTS idx_audit_logs_path ON audit_logs(path);

