CREATE TABLE IF NOT EXISTS schools (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  logo_url text,
  address text,
  phone text,
  email text,
  website text,
  principal_name text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- Add school_id to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS school_id uuid REFERENCES schools(id);

-- Create a default school and link existing users
DO $$
DECLARE
    default_school_id uuid;
    s_name text;
    s_logo text;
BEGIN
    -- Try to get current identity from settings
    SELECT value_json->>'school_name', value_json->>'logo_url' 
    INTO s_name, s_logo
    FROM app_settings WHERE key = 'school_identity';

    IF s_name IS NULL OR s_name = '' THEN
        s_name := 'AtigaCBT Default School';
    END IF;

    -- Insert into schools
    INSERT INTO schools (name, logo_url) 
    VALUES (s_name, COALESCE(s_logo, '')) 
    RETURNING id INTO default_school_id;

    -- Update existing users to this school
    UPDATE users SET school_id = default_school_id WHERE school_id IS NULL;
END $$;
