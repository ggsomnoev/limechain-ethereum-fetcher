BEGIN;

DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS users;

INSERT INTO users (username, password) VALUES
('alice', 'alice'),
('bob', 'bob'),
('carol', 'carol'),
('dave', 'dave');

COMMIT;