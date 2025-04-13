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

COMMIT;