ALTER TABLE registration_requests
  ADD COLUMN IF NOT EXISTS password_hash text NULL,
  ADD COLUMN IF NOT EXISTS nis text NULL,
  ADD COLUMN IF NOT EXISTS nip text NULL,
  ADD COLUMN IF NOT EXISTS program_code text NULL,
  ADD COLUMN IF NOT EXISTS level_name text NULL,
  ADD COLUMN IF NOT EXISTS group_name text NULL,
  ADD COLUMN IF NOT EXISTS mapel_codes text NULL;

