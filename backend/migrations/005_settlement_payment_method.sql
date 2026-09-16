-- Optional payment-app method recorded when the payer confirms after opening Venmo/etc.
ALTER TABLE settlements
ADD COLUMN IF NOT EXISTS payment_method text NOT NULL DEFAULT ''
CHECK (payment_method IN ('', 'venmo', 'paypal', 'cashApp', 'zelle', 'other'));
