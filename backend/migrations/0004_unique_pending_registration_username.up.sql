CREATE UNIQUE INDEX IF NOT EXISTS uniq_registration_pending_username
  ON registration_requests (username)
  WHERE status = 'pending';

