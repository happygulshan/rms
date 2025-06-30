CREATE TABLE IF NOT EXISTS users  (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, --add later TIMESTAMPTZ
    archived_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_active_emails ON users(TRIM(LOWER(email))) WHERE archived_at IS NULL;
