ALTER TABLE registration_requests
  ADD COLUMN IF NOT EXISTS phone text NULL;
