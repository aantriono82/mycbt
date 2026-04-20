CREATE TABLE IF NOT EXISTS login_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NULL REFERENCES users(id) ON DELETE SET NULL,
  username text NOT NULL,
  role text NOT NULL,
  ip text NOT NULL,
  user_agent text NOT NULL DEFAULT '',
  logged_in_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_login_logs_logged_in_at ON login_logs(logged_in_at DESC);
CREATE INDEX IF NOT EXISTS idx_login_logs_username ON login_logs(username);

