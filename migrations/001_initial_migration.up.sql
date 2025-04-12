BEGIN;

CREATE TABLE IF NOT EXISTS transactions (
    transaction_hash TEXT PRIMARY KEY,  
    transaction_data JSONB NOT NULL,    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;