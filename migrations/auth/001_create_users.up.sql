-- +migrate Up
-- SQL in this section is executed when the migration is applied (e.g., creating the table).

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(64) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create an index on email for faster lookups (optional but recommended)
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);