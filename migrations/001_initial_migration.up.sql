BEGIN;

CREATE TABLE IF NOT EXISTS transactions (
    transaction_hash TEXT PRIMARY KEY,  
    transaction_data JSONB NOT NULL,    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  password TEXT NOT NULL
);
-- for the sake of testing we will store passwords in plain text
INSERT INTO users (username, password) VALUES
('alice', 'alice'),
('bob', 'bob'),
('carol', 'carol'),
('dave', 'dave')
ON CONFLICT (username) DO NOTHING;

CREATE TABLE IF NOT EXISTS auth_tokens (
    token TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL
);

COMMIT;