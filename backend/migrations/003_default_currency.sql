-- Change the preferred_currency column default to USD.
-- This only affects brand-new rows inserted without an explicit value.
-- Existing rows retain their current currency; do NOT add an UPDATE here.
ALTER TABLE users ALTER COLUMN preferred_currency SET DEFAULT 'USD';
