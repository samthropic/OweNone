CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL,
    display_name text NOT NULL CHECK (length(trim(display_name)) BETWEEN 1 AND 100),
    preferred_currency text NOT NULL DEFAULT 'GBP' CHECK (preferred_currency ~ '^[A-Z]{3}$'),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_email_unique ON users (lower(email));

CREATE TABLE waitlist_entries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX waitlist_entries_email_unique ON waitlist_entries (lower(email));

CREATE TABLE friendships (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, friend_id),
    CHECK (user_id < friend_id)
);

CREATE TABLE groups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL CHECK (length(trim(name)) BETWEEN 1 AND 100),
    icon text NOT NULL DEFAULT '👥' CHECK (length(icon) BETWEEN 1 AND 16),
    created_by uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE group_members (
    group_id uuid NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role text NOT NULL DEFAULT 'member' CHECK (role IN ('owner', 'member')),
    joined_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX group_members_user_id_idx ON group_members (user_id, group_id);

CREATE TABLE expenses (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id uuid NOT NULL REFERENCES groups(id),
    description text NOT NULL CHECK (length(trim(description)) BETWEEN 1 AND 200),
    category text NOT NULL DEFAULT 'general' CHECK (length(category) BETWEEN 1 AND 40),
    amount_minor bigint NOT NULL CHECK (amount_minor > 0),
    currency text NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    paid_by uuid NOT NULL REFERENCES users(id),
    split_method text NOT NULL CHECK (split_method IN ('equal', 'exact')),
    created_by uuid NOT NULL REFERENCES users(id),
    idempotency_key text NOT NULL CHECK (length(idempotency_key) BETWEEN 8 AND 200),
    request_fingerprint text NOT NULL CHECK (length(request_fingerprint) = 64),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (created_by, idempotency_key)
);

CREATE INDEX expenses_group_created_at_idx ON expenses (group_id, created_at DESC);

CREATE TABLE expense_splits (
    expense_id uuid NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id),
    amount_minor bigint NOT NULL CHECK (amount_minor > 0),
    PRIMARY KEY (expense_id, user_id)
);

CREATE OR REPLACE FUNCTION validate_expense_split_total()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    target_expense_id uuid := COALESCE(NEW.expense_id, OLD.expense_id);
    expected_total bigint;
    actual_total bigint;
BEGIN
    SELECT amount_minor INTO expected_total FROM expenses WHERE id = target_expense_id;
    IF NOT FOUND THEN
        RETURN NULL;
    END IF;

    SELECT COALESCE(sum(amount_minor), 0) INTO actual_total
    FROM expense_splits
    WHERE expense_id = target_expense_id;

    IF actual_total <> expected_total THEN
        RAISE EXCEPTION 'expense % splits total %, expected %', target_expense_id, actual_total, expected_total;
    END IF;
    RETURN NULL;
END;
$$;

CREATE CONSTRAINT TRIGGER expense_splits_total
AFTER INSERT OR UPDATE OR DELETE ON expense_splits
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION validate_expense_split_total();

CREATE TABLE settlements (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id uuid REFERENCES groups(id),
    from_user_id uuid NOT NULL REFERENCES users(id),
    to_user_id uuid NOT NULL REFERENCES users(id),
    amount_minor bigint NOT NULL CHECK (amount_minor > 0),
    currency text NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    status text NOT NULL DEFAULT 'completed' CHECK (status IN ('pending', 'completed', 'failed')),
    idempotency_key text NOT NULL CHECK (length(idempotency_key) BETWEEN 8 AND 200),
    request_fingerprint text NOT NULL CHECK (length(request_fingerprint) = 64),
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (from_user_id <> to_user_id),
    UNIQUE (from_user_id, idempotency_key)
);

CREATE INDEX settlements_participants_created_at_idx
ON settlements (from_user_id, to_user_id, created_at DESC)
WHERE status = 'completed';

CREATE TABLE reminders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id uuid NOT NULL REFERENCES users(id),
    recipient_id uuid NOT NULL REFERENCES users(id),
    message text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (sender_id <> recipient_id)
);

CREATE INDEX reminders_sender_created_at_idx ON reminders (sender_id, created_at DESC);