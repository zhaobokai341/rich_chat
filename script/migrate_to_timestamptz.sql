-- Migration script to update existing database schema
-- This script converts TIMESTAMP to TIMESTAMPTZ and adds performance indexes
-- Run this ONCE on your existing rich_chat database

\c rich_chat;

-- Set timezone to UTC for consistent handling
SET TIME ZONE 'UTC';

-- Convert timestamp columns to timestamptz (timestamp with timezone)
-- This preserves existing data and interprets it as UTC
ALTER TABLE users 
    ALTER COLUMN registered_time TYPE TIMESTAMPTZ USING registered_time AT TIME ZONE 'UTC',
    ALTER COLUMN last_login TYPE TIMESTAMPTZ USING last_login AT TIME ZONE 'UTC',
    ALTER COLUMN lock_until TYPE TIMESTAMPTZ USING lock_until AT TIME ZONE 'UTC';

ALTER TABLE ip_block_list
    ALTER COLUMN block_until TYPE TIMESTAMPTZ USING block_until AT TIME ZONE 'UTC',
    ALTER COLUMN created_time TYPE TIMESTAMPTZ USING created_time AT TIME ZONE 'UTC';

-- Add indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_lock_until ON users(lock_until) WHERE lock_until IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_ip_block_list_ip ON ip_block_list(ip);
CREATE INDEX IF NOT EXISTS idx_ip_block_list_block_until ON ip_block_list(block_until);

-- Verify the changes
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'users' 
AND column_name IN ('registered_time', 'last_login', 'lock_until');

SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'ip_block_list' 
AND column_name IN ('block_until', 'created_time');

-- List all indexes
SELECT indexname, tablename 
FROM pg_indexes 
WHERE schemaname = 'public' 
AND tablename IN ('users', 'ip_block_list')
ORDER BY tablename, indexname;