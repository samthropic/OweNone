-- Linked payment-app handles for faster settle-up.
-- Empty string is stored as NULL so "not linked" stays distinct from a handle.
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS venmo_handle   text,
  ADD COLUMN IF NOT EXISTS paypal_handle  text,
  ADD COLUMN IF NOT EXISTS cashapp_handle text,
  ADD COLUMN IF NOT EXISTS zelle_handle   text;

ALTER TABLE users
  ADD CONSTRAINT users_venmo_handle_len
    CHECK (venmo_handle IS NULL OR char_length(venmo_handle) BETWEEN 1 AND 64),
  ADD CONSTRAINT users_paypal_handle_len
    CHECK (paypal_handle IS NULL OR char_length(paypal_handle) BETWEEN 1 AND 64),
  ADD CONSTRAINT users_cashapp_handle_len
    CHECK (cashapp_handle IS NULL OR char_length(cashapp_handle) BETWEEN 1 AND 64),
  ADD CONSTRAINT users_zelle_handle_len
    CHECK (zelle_handle IS NULL OR char_length(zelle_handle) BETWEEN 1 AND 128);
