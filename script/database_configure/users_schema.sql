CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(50) DEFAULT '',
    nickname VARCHAR(50) DEFAULT '',
    bio TEXT DEFAULT '',
    password_hash VARCHAR(255) NOT NULL,
    registered_time TIMESTAMPTZ DEFAULT NOW(),
    last_login TIMESTAMPTZ NOT NULL,
    lock_until TIMESTAMPTZ -- NULL means unlocked, timestamp with timezone means locked until this time
);

-- Add indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_lock_until ON users(lock_until) WHERE lock_until IS NOT NULL;
