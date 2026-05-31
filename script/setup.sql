DROP DATABASE IF EXISTS rich_chat;
CREATE DATABASE rich_chat;

\c rich_chat;

-- Set default timezone to UTC for consistent timestamp handling
SET TIME ZONE 'UTC';

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    nickname VARCHAR(50) NOT NULL,
    bio TEXT NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    registered_time TIMESTAMPTZ DEFAULT NOW(),
    last_login TIMESTAMPTZ NOT NULL,
    lock_until TIMESTAMPTZ -- NULL means unlocked, timestamp with timezone means locked until this time
);

-- Add indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_lock_until ON users(lock_until) WHERE lock_until IS NOT NULL;

CREATE TABLE IF NOT EXISTS ip_block_list (
    ip VARCHAR(50) NOT NULL,
    block_until TIMESTAMPTZ NOT NULL,
    reason TEXT,
    created_time TIMESTAMPTZ DEFAULT NOW()
);

-- Add index for IP blocking lookups
CREATE INDEX IF NOT EXISTS idx_ip_block_list_ip ON ip_block_list(ip);
CREATE INDEX IF NOT EXISTS idx_ip_block_list_block_until ON ip_block_list(block_until);