ALTER TABLE users ADD COLUMN IF NOT EXISTS google_id TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id) WHERE google_id IS NOT NULL;

ALTER TABLE registration_requests ADD COLUMN IF NOT EXISTS google_id TEXT;
ALTER TABLE registration_requests ADD COLUMN IF NOT EXISTS google_picture TEXT;
