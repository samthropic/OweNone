-- Profile photos visible to friends and group mates.
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS avatar_url text;

ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_avatar_url_len;

ALTER TABLE users
  ADD CONSTRAINT users_avatar_url_len
    CHECK (avatar_url IS NULL OR char_length(avatar_url) BETWEEN 1 AND 512);
