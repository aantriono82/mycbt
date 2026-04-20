ALTER TABLE registration_requests
  DROP COLUMN IF EXISTS password_hash,
  DROP COLUMN IF EXISTS nis,
  DROP COLUMN IF EXISTS nip,
  DROP COLUMN IF EXISTS program_code,
  DROP COLUMN IF EXISTS level_name,
  DROP COLUMN IF EXISTS group_name,
  DROP COLUMN IF EXISTS mapel_codes;

