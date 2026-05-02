ALTER TABLE registration_requests
ADD COLUMN IF NOT EXISTS password_plain text;
